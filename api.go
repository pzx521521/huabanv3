package huabanv3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

type UploadOptions struct {
	//上传失败是否打断上传
	Break bool
	//上传失败后的重试次数
	ReUploadTime int
	//自定义标签
	TagFunc func(path string) []string
	//自定义Argc
	ArgcFunc func(path string) *Aigc
	//自定义Title
	TitleFunc func(path string) string
}
type HuaBanAPIV3 struct {
	account   string
	password  string
	boardName string
	userName  string
	client    *http.Client
	boards    *BoardsResponse
	boardID   int
	// 默认每次传多少个(即网页端的限制50个)
	BatchSize int
	// 每次多少个同时上传(上传这50个时.每次上传多少个)
	// 就是网页上传进度条同时跑的数量
	PoolSize       int
	UploadOptions  *UploadOptions
	FailFiles      []string
	FailBoardFiles []string
	SuccessFiles   map[string]*File
	Header         map[string]string
}

func NewHuaBanApiV3(account, password string) *HuaBanAPIV3 {
	return &HuaBanAPIV3{
		account:      account,
		password:     password,
		client:       http.DefaultClient,
		BatchSize:    50,
		PoolSize:     10,
		SuccessFiles: make(map[string]*File),
		UploadOptions: &UploadOptions{
			Break:        false,
			ReUploadTime: 2,
		},
		Header: map[string]string{"User-Agent": UA},
	}
}

func (hu *HuaBanAPIV3) SetClient(client *http.Client) {
	hu.client = client
}
func (hu *HuaBanAPIV3) Login() error {
	userName, err := Login(hu.client, hu.account, hu.password, hu.Header)
	if err != nil {
		log.Printf("获取用户名失败...%v\n", err)
		return err
	}
	hu.userName = userName
	//获取Board
	boards, err := GetBoards(hu.client, hu.Header, hu.userName)
	hu.boards = boards
	if err != nil {
		log.Printf("获取画板信息失败...%v\n", err)
		return err
	}
	return nil
}
func (hu *HuaBanAPIV3) CheckNotExist(files []string, boardName string) ([]string, error) {
	board, err := hu.GetBoard(boardName)
	if err != nil {
		return nil, err
	}
	infos, _ := hu.GetImgInfos(board.BoardId)
	m := map[string]string{}
	for _, file := range files {
		fileName := filepath.Base(file)
		//fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
		m[fileName] = file
	}

	for _, info := range infos {
		_, ok := m[info.RawText]
		if ok {
			m[info.RawText] = ""
		}
	}
	var notExitFile []string
	for key := range m {
		if m[key] != "" {
			notExitFile = append(notExitFile, m[key])
		}
	}
	return notExitFile, nil
}

// UploadBatch 批量上传文件到指定的画板。
// 如果 boardName 为空，则只上传文件而不添加到任何画板。
// 参数:
//
//	files - 待上传的文件路径列表。
//	boardName - 要添加文件的画板名称。
func (hu *HuaBanAPIV3) UploadBatch(files []string, boardName string) error {
	//允许只上传不添加到画板
	if boardName != "" {
		hu.boardName = boardName
		board, err := hu.GetBoard(boardName)
		if err != nil {
			return err
		}
		hu.boardID = board.BoardId
	}
	return hu.UploadBatchByBoardID(files, hu.boardID)
}
func (hu *HuaBanAPIV3) UploadBatchByBoardID(files []string, boardID int) error {
	//允许只上传不添加到画板
	hu.boardID = boardID
	for i := 0; i < len(files); i += hu.BatchSize {
		end := i + hu.BatchSize
		if end > len(files) {
			end = len(files)
		}
		batch := files[i:end]
		err := hu.uploadBatch(batch)
		if err != nil {
			log.Printf("goutine:%d批量上传失败:%v\n以下文件未能添加到画板%s\n", i, err, batch)
			hu.FailBoardFiles = append(hu.FailBoardFiles, batch...)
			return err
		}
	}
	return nil
}

func (hu *HuaBanAPIV3) CreateBoard(boardName string) (*Board, error) {
	return createBoard(hu.client, hu.Header, boardName)
}

func (hu *HuaBanAPIV3) GetBoard(boardName string) (*Board, error) {
	board, err := GetBoard(hu.client, hu.Header, hu.boards, boardName)
	if err != nil {
		return nil, err
	}
	return board, nil
}
func (hu *HuaBanAPIV3) reUpload(groutineID int, file string) (*File, error) {
	log.Printf("goutine:%d继续%d次重试上传\n", groutineID, hu.UploadOptions.ReUploadTime)
	for time := 0; time < hu.UploadOptions.ReUploadTime; time++ {
		fileInfo, err := upload(hu.client, hu.Header, file)
		if err == nil {
			return fileInfo, nil
		}
		log.Printf("goutine:%d重试第%d次上传失败\n", groutineID, time)
	}
	hu.FailFiles = append(hu.FailFiles, file)
	return nil, errors.New(fmt.Sprintf("goutine:%d重试次数用完，放弃上传\n", groutineID))
}

func (hu *HuaBanAPIV3) reAddBoard(client *http.Client, body *BatchInfo) error {
	log.Printf("开始一共%d次的添加到画板重试\n", hu.UploadOptions.ReUploadTime)
	for time := 0; time < hu.UploadOptions.ReUploadTime; time++ {
		err := addBoard(client, hu.Header, body)
		if err == nil {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("goutine:%d重试次数用完，放弃添加到画板\n", hu.UploadOptions.ReUploadTime))
}
func (hu *HuaBanAPIV3) upload(files []string) (map[string]*File, error) {
	eg := errgroup.Group{}
	//用于控制并发的大小
	ch := make(chan struct{}, hu.PoolSize)
	ret := map[string]*File{}
	var lock sync.Mutex
	for i, file := range files {
		eg.Go(func() error {
			ch <- struct{}{}
			defer func() {
				<-ch
			}()
			index := i // 重新声明 i 为局部变量
			filePath := file
			fileInfo, err := upload(hu.client, hu.Header, filePath)
			if err != nil || fileInfo == nil {
				log.Printf("goutine:%d上传失败了:%v, 路径%s\n", index, err, file)
				hu.FailFiles = append(hu.FailFiles, file)
				if errors.Is(err, &UploadIgnoreError{}) {
					return nil
				}
				return err
			}
			lock.Lock()
			hu.SuccessFiles[filePath] = fileInfo
			ret[file] = fileInfo
			lock.Unlock()
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (hu *HuaBanAPIV3) uploadBatch(files []string) error {
	batchInfo := &BatchInfo{BoardId: hu.boardID}
	fileInfos, err := hu.upload(files)
	if err != nil {
		return err
	}
	pins := []*UploadPin{}
	for filePath, fileInfo := range fileInfos {
		tags := []string{}
		var argc *Aigc
		title := filepath.Base(filePath)
		if hu.UploadOptions.TagFunc != nil {
			tags = hu.UploadOptions.TagFunc(filePath)
		}
		if hu.UploadOptions.ArgcFunc != nil {
			argc = hu.UploadOptions.ArgcFunc(filePath)
		}
		if hu.UploadOptions.TitleFunc != nil {
			title = hu.UploadOptions.TitleFunc(filePath)
		}
		pins = append(pins,
			&UploadPin{
				FileId: fileInfo.Id,
				Text:   title,
				Tags:   tags,
				Aigc:   argc,
			})
	}
	if len(pins) == 0 {
		return errors.New("没有文件可以上传")
	}
	batchInfo.Pins = pins
	err = addBoard(hu.client, hu.Header, batchInfo)
	if err != nil {
		return err
	}
	return nil
}

func (hu *HuaBanAPIV3) ChangeTags(info *PutPinInfo) error {
	//'{"pin_id":6375141434,"board_id":94018210,"text":"灰黑红805f11","link":"213123","tags":["测试长度的标签"],"aigc":{"aigc_category":"stable_diffusion","model":"123123123","prompt":"11111111"}}'
	if info == nil || info.PinId == 0 || info.BoardId == 0 {
		return nil
	}
	apiUrl := fmt.Sprintf("%s/v3/pins/%d", DOMAIN, info.PinId)
	jsonData, err := json.Marshal(info)
	// 创建请求主体
	req, err := http.NewRequest("PUT", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	SetHeader(req, hu.Header)
	SetHeaderAsJson(req)
	res, err := hu.client.Do(req)
	if err != nil {
		return errors.Join(errors.New("修改信息失败: "), err)
	}

	var pinResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&pinResponse); err != nil {
		return err
	}
	log.Printf("修改信息成功%v\n", pinResponse)
	return nil
}

func (hu *HuaBanAPIV3) GetImgInfos(boardID int) ([]*Pin, error) {
	return GetImgInfos(hu.client, boardID, hu.Header)
}
