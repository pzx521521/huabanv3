package huabanv3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 代理 通过不同ip下载/抓包
func GetProxyHttpClient(porxyUrl string) *http.Client {
	// 创建代理 URL
	if porxyUrl == "" {
		porxyUrl = "http://localhost:7897"
	}
	// clash 代理
	//proxyURL, _ := url.Parse("http://localhost:7897")
	//charles 代理
	//proxyURL, _ := url.Parse("http://localhost:8888")
	proxyURL, _ := url.Parse(porxyUrl)
	// 创建一个带有代理的 Transport
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	// 创建一个带有自定义 Transport 的 Client
	client := &http.Client{
		Transport: transport,
	}
	return client
}

// 获取登录 Cookie 的函数
func getCookie(client *http.Client, account, password string) (string, error) {
	apiUrl := DOMAIN + "/v3/auth/"
	dataUrl := url.Values{
		"email":    {account},
		"password": {password},
	}
	data := bytes.NewBufferString(dataUrl.Encode())
	// 创建一个 HTTP 客户端，并设置 CheckRedirect 函数来禁止重定向

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer func() {
		client.CheckRedirect = nil
	}()
	req, err := http.NewRequest("POST", apiUrl, data)
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return "", err
	}
	SetHeaderUA(req)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	// 执行请求
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return "", err
	}
	defer res.Body.Close()
	// 读取响应数据
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return "", err
	}
	// 处理响应
	responseText := string(body)
	if !strings.HasPrefix(responseText, "Found.") {
		msg := fmt.Sprintf("用户名密码登录失败:%v\n", responseText)
		return "", errors.New(msg)
	}

	// 提取 Cookie
	cookies := res.Cookies()
	cookieArr := []string{}
	for _, cookie := range cookies {
		cookieArr = append(cookieArr, cookie.Name+"="+cookie.Value)
	}
	cookieStr := strings.Join(cookieArr, ";")
	if cookieStr == "" {
		fmt.Println("登录失败，未知原因")
		fmt.Println("响应内容:", responseText)
	}
	return cookieStr, nil
}

func deleteBoard(client *http.Client, BoardId string) error {
	url := "https://huaban.com/v3/boards/" + BoardId
	method := "DELETE"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	// 发送请求
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}
func createBoard(client *http.Client, headers map[string]string, title string) (*Board, error) {
	apiUrl := DOMAIN + "/v3/boards"
	payload := map[string]string{
		"title":      title, //中文好难搞
		"is_private": "0",
		"creation":   "false",
	}
	marshal, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	// 创建请求
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(marshal))
	SetHeader(req, headers)
	SetHeaderAsJson(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	boardResponse := &BoardResponse{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &boardResponse)
	if err != nil {
		return nil, err
	}
	if boardResponse.Board == nil {
		return nil, errors.New(string(body))
	}
	return boardResponse.Board, nil
}
func GetBoard(client *http.Client, headers map[string]string, boards *BoardsResponse, boardName string) (board *Board, err error) {
	board = boards.ExistBoard(boardName)
	if board != nil {
		return board, nil
	} else {
		log.Printf("画板不存在，开始创建画板\n")
		board, err = createBoard(client, headers, boardName)
		if err != nil {
			log.Printf("创建画板失败...%v\n", err)
			return
		}
		log.Printf("创建画板成功，开始上传图片\n")
	}
	return
}

// 获取用户数据
func GetBoards(client *http.Client, headers map[string]string, username string) (*BoardsResponse, error) {
	apiUrl := DOMAIN + "/v3/" + username + "/boards?limit=30"
	req, _ := http.NewRequest("GET", apiUrl, nil)
	SetHeader(req, headers)
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Join(errors.New("获取用户主页信息失败:"), err)
	}
	defer res.Body.Close()
	ret := &BoardsResponse{}
	err = json.NewDecoder(res.Body).Decode(ret)
	if err != nil {
		return nil, errors.Join(errors.New("解析用户主页信息失败:"), err)
	}
	return ret, nil
}

// 获取urlname
func getUserName(client *http.Client, headers map[string]string) (username string, err error) {
	apiUrl := DOMAIN + "/follow"
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return
	}
	SetHeader(req, headers)
	// 执行请求
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	// 读取响应数据
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	responseText := string(body)

	// 匹配用户名
	usernameRegexp := regexp.MustCompile(`"urlname":"(.+?)"`)
	matches := usernameRegexp.FindStringSubmatch(responseText)
	if matches != nil && len(matches) > 1 {
		username = matches[1]
		return username, nil
	}

	return "", errors.New("找不到urlname")
}

func getMutipart(filepath string) (*bytes.Buffer, *multipart.Writer, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	// 将图像数据作为普通的字段直接添加到 multipart 数据中
	part, err := writer.CreateFormFile("file", "")
	if err != nil {
		return nil, nil, err
	}
	// 将 imageData 写入该部分
	io.Copy(part, file)
	// 结束 multipart 写入
	writer.Close()
	return &b, writer, nil
}

// 添加画板
func addBoard(client *http.Client, header map[string]string, body *BatchInfo) error {
	if body.BoardId == 0 {
		log.Printf("board id:%d, 不进行添加到画板操作\n", body.BoardId)
		return nil
	}
	apiUrl := DOMAIN + "/v3/pins/batch"
	jsonData, err := json.Marshal(body)
	// 创建请求主体
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	SetHeader(req, header)
	SetHeaderAsJson(req)
	res, err := client.Do(req)
	if err != nil {
		return errors.Join(errors.New("添加图片文件到画板失败: "), err)
	}

	var pinResponse map[string]interface{}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &pinResponse); err != nil {
		return err
	}
	if _, ok := pinResponse["result"]; !ok {
		return errors.New(fmt.Sprintf("添加图片到画板%d失败:%s\n", body.BoardId, data))
	}
	a := pinResponse["result"].(map[string]interface{})
	log.Printf("成功添加%d图片到画板%d\n", len(a), body.BoardId)
	return nil
}

// 上传单个文件
func upload(client *http.Client, header map[string]string, filePath string) (fileInfo *File, err error) {
	apiUrl := DOMAIN + "/v3/upload"
	// 创建一个缓冲区和multipart writer
	body, writer, err := getMutipart(filePath)
	if err != nil {
		return
	}
	// 创建请求
	req, err := http.NewRequest("POST", apiUrl, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	SetHeader(req, header)
	if err != nil {
		return
	}
	// 发送请求
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &fileInfo)
	if err != nil {
		return nil, err
	}
	if fileInfo.Key == "" {
		return nil, errors.New(string(data))
	}
	log.Printf("上传成功:%s, 地址:%s", filePath, fileInfo.Key)
	return fileInfo, nil
}

func GetAllFiles(path string) ([]string, error) {
	var files []string
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		files = append(files, path)
		return files, err
	}
	err = filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".jpg", ".png", ".jpeg", ".webp", ".gif", "bmp":
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func GetImgUrl(f *File) string {
	return fmt.Sprintf("https://%s.huaban.com/%s",
		f.Bucket, f.Key)
}
