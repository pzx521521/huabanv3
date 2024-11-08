package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/pzx521521/huabanv3"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

const CONFIGPATH = "./config.json"

type Config struct {
	Name  string   `json:"name"`
	Pass  string   `json:"pass"`
	Board string   `json:"board"`
	Dir   []string `json:"dir"`
	Debug bool     `json:"debug"`
}

func main() {
	config, err := ParseJson()
	if err != nil {
		config, err = ParseFlag()
		if err != nil {
			log.Printf("%v", err)
			return
		}
	}

	if !config.Debug {
		log.SetOutput(io.Discard)
	}
	var allFile []string

	//获取文件
	for _, filePath := range config.Dir {
		files, err := huabanv3.GetAllFiles(filePath)
		if err != nil {
			log.Printf("获取硬盘图片文件:%s失败...%v\n", files, err)
		}
		allFile = append(allFile, files...)
	}
	if len(allFile) == 0 {
		log.Printf("未找到图片文件\n")
		return
	}
	//获取cookie
	log.Printf("你的设置: 账号: %s, 密码: %s, 图片目录路径: %s, 画板名: %s\n",
		config.Name, config.Pass, config.Dir, config.Board)
	huaBanApi := huabanv3.NewHuaBanApiV3(config.Name, config.Pass)
	client := http.DefaultClient
	huaBanApi.SetClient(client)
	err = huaBanApi.Login()
	if err != nil {
		log.Printf("登录失败...%v\n", err)
		return
	}

	//上传
	err = huaBanApi.UploadBatch(allFile, config.Board)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	//结果
	marshal, err := json.Marshal(huaBanApi.SuccessFiles)
	if err != nil {
		return
	}
	log.Printf("%s\n", marshal)
	for _, sf := range huaBanApi.SuccessFiles {
		imgUrl := fmt.Sprintf("https://%s.huaban.com/%s",
			sf.Bucket, sf.Key)
		fmt.Println(imgUrl)
	}
}

func ParseJson() (*Config, error) {
	absPath := path.Join(path.Dir(os.Args[0]), CONFIGPATH)
	var config Config
	// 检查 cookie.json 文件是否存在且有效
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		return nil, err
	}
	content, err := os.ReadFile(absPath)
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	if len(os.Args) > 1 {
		config.Dir = os.Args[1:]
	}
	return &config, nil
}
func ParseFlag() (*Config, error) {
	// 定义带名字的参数
	account := flag.String("name", "", "账户名")
	password := flag.String("pass", "", "账户密码")
	boardName := flag.String("board", "", "画板名")
	dirPath := flag.String("dir", "", "图片文件目录路径")
	debug := flag.Bool("debug", false, "日志输出")
	// 解析命令行参数
	flag.Parse()
	// 检查账号和密码
	if *account == "" || *password == "" {
		return nil, errors.New("未设定账密")
	}
	// 检查图片目录路径
	if *dirPath == "" {
		return nil, errors.New("未设定图片目录路径")
	}
	// 检查画板名
	if *boardName == "" {
		return nil, errors.New("未设定画板名")
	}
	return &Config{*account, *password, *boardName, []string{*dirPath}, *debug}, nil
}
