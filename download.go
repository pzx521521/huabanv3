package huabanv3

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// DownloadBoard 下载指定boardID的所有图片。
// 参数:
//
//	boardID - 需要下载的board的ID。
//	poolSize - 并发下载的最大数量。
//	rawText - 是否使用网页中的描述字段作为图片的名称。为了防止文件名非法，默认使用图片的ID为文件名。
func DownloadBoard(boardID int, poolSize int, rawText bool) error {
	dirName := filepath.Join("./download/", strconv.Itoa(boardID))
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return err
	}

	var eg errgroup.Group
	semaphore := make(chan struct{}, poolSize)
	client := http.DefaultClient
	pinInfo, err := GetImgInfos(client, boardID)
	log.Printf("共发现%v张图片\n", len(pinInfo))
	if err != nil {
		return err
	}
	for index, imageInfo := range pinInfo {
		info := imageInfo
		index := index
		eg.Go(func() error {
			//return err 不继续执行
			//return nil 继续执行
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()
			err := downloadImageMuti(client, info, dirName, rawText)

			if err != nil {
				log.Printf("下载%s错处%v\n", info.RawText, err)
			}
			log.Printf("%d开始下载:%s ", index, info.RawText)
			return nil
		})
	}
	err = eg.Wait()
	if err != nil {
		return err
	}
	log.Printf("所有图片下载完成:%d\n", boardID)
	return nil
}
func downloadImageMuti(client *http.Client, imgInfo *Pin, dirPath string, rawText bool) error {
	byType, err := mime.ExtensionsByType(imgInfo.File.Type)
	if err != nil {
		return err
	}
	filePath := filepath.Join(dirPath, imgInfo.File.Key+byType[len(byType)-1])
	if rawText {
		filePath = filepath.Join(dirPath, imgInfo.RawText)
	}
	_, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		return nil
	}
	imgUrl := fmt.Sprintf("https://%s.huaban.com/%s", imgInfo.File.Bucket, imgInfo.File.Key)
	// 发送 HTTP 请求获取图片内容
	resp, err := client.Get(imgUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	// 创建文件
	err = os.WriteFile(filePath, data, 0644)
	log.Printf("下载完成:%v\n", filePath)
	if err != nil {
		return err
	}
	return nil
}

func GetImgInfos(client *http.Client, boardID int) ([]*Pin, error) {
	const limit = 100
	pinsResponse, err := GetPins(client, boardID, 1, 0)
	if err != nil {
		return nil, err
	}
	if pinsResponse.Board == nil {
		return nil, errors.New("没有找到board")
	}
	pinCount := pinsResponse.Board.PinCount
	log.Printf("%d有%d张图片\n", boardID, pinCount)
	if pinsResponse.Pins == nil || len(pinsResponse.Pins) == 0 {
		return nil, errors.New("没有图片")
	}
	pinId := pinsResponse.Pins[0].PinID
	ret := []*Pin{}
	ret = append(ret, pinsResponse.Pins[0])
	pageCount := (pinCount + limit - 1) / 100
	for i := 0; i < pageCount; i++ {
		pinsResponse, err := GetPins(client, boardID, 100, pinId)
		if err != nil {
			return nil, err
		}
		if pinsResponse.Pins == nil || len(pinsResponse.Pins) == 0 {
			break
		}
		for _, pin := range pinsResponse.Pins {
			ret = append(ret, pin)
		}
		pinId = pinsResponse.Pins[len(pinsResponse.Pins)-1].PinID
	}
	return ret, nil
}

// 获取画板内的图片
func GetPins(client *http.Client, boardID, limit int, pinId int64) (*PinsResponse, error) {
	sPinID := ""
	if pinId > 0 {
		sPinID = strconv.FormatInt(pinId, 10)
	}
	apiUrl := DOMAIN + fmt.Sprintf(
		"/v3/boards/%d/pins?limit=%d&max=%s&fields=pins:PIN",
		boardID, limit, sPinID)

	//专门用于查询有多少张图片
	if limit == 1 {
		//|Cboard:BOARD_DETAIL
		apiUrl += "%7Cboard:BOARD_DETAIL"
	}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	SetHeaderUA(req)
	// 发送请求
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	pins := PinsResponse{}
	err = json.NewDecoder(res.Body).Decode(&pins)
	if err != nil {
		return nil, err
	}
	return &pins, nil
}
