# [花瓣网图片 API v3](https://github.com/pzx521521/huabanv3) golang版本
花瓣网图片批量上传/下载工具  
[python版本的](https://github.com/Pingze-github/HuabanBatchUpload)这个太老了 已经不能用了,现在是V3接口  
可以作为**图床**.  
[示例程序](bahttps://github.com/pzx521521/typora-plugin-huan)
[github](https://github.com/pzx521521/huabanv3)
## 与[bilibili](https://github.com/xlzy520/bilibili-img-uploader)图床比较
|          | 跨域 | 网页直接使用 | 不压缩          |
|----------|---|--------|--------------|
| bilibili | ✅ | ❌(需要no-referrer)  | ❌(部分压缩机制不知道) |
| huaban   | ❌ | ✅      | ✅            |

bilibili跨域很好玩,比如在图片中存储数据(写在最后面),然后fetch 获取数据

## 功能说明
+ 无需验证的
  + 批量下载图片
    + 错误自动重试
    + 多线程并发数量设置
    + 使用描述作为文件名(未对描述做任何处理,特殊字符可能导致写入失败)
  + 获取画板图片列表
+ 需要验证的
    + 批量上传
      + 每张图片 自定义tag/argc, tags 可以用 [azure-vision](https://github.com/pzx521521/azurevision) 的api 生成  
      + 多线程并发，可以快速上传大量文件,添加错误重试机制
    + 新建删除画板
    + 修改图片的tag/argc/描述 描述可以通过[azure-vision](https://github.com/pzx521521/azurevision) 的api 生成

## 使用方法 
### 如果你只想要上传功能 
使用[typora-plugin-huaban](https://github.com/pzx521521/typora-plugin-huaban)  

### 安装
```cmd
go get github.com/pzx521521/huabanv3
```

### 下载图片 无需验证的
```go
package main
import "github.com/pzx521521/huabanv3"

func main()  {
	huabanv3.DownloadBoard(94004345, 10, false)
}
```

### 上传图片 需要验证的
```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/pzx521521/huabanv3"
	"log"
)

type Config struct {
	Name  string `json:"name"`
	Pass  string `json:"pass"`
	Board string `json:"board"`
	Dir   string `json:"dir"`
	Debug bool   `json:"debug"`
}
func main() {
	config := &Config{
		Name:  "your_name",
		Pass:  "",
		Board: "your_board_name",
		Dir:   "your_dir",
		Debug: true,
	}
	//获取cookie
	huaBanApi := huabanv3.NewHuaBanApiV3(config.Name, config.Pass)
	err := huaBanApi.Login()
	if err != nil {
		log.Printf("登录失败...%v\n", err)
		return
	}
	//获取文件 可以是文件/也可以是文件夹
	files, err := huabanv3.GetAllFiles(config.Dir)
	if err != nil {
		log.Printf("获取硬盘图片文件失败...%v\n", err)
		return
	}
	//上传
	err = huaBanApi.UploadBatch(files, config.Board)
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

	//单路径结果
	for _, sf := range huaBanApi.SuccessFiles {
		imgUrl := fmt.Sprintf("https://%s.huaban.com/%s",
			sf.Bucket, sf.Key)
		fmt.Println(imgUrl)
	}
}
```

### 自定义每一张图片的title/tag/argc
```go
package main

import "github.com/pzx521521/huabanv3"

func main() {
	huaBanApi := huabanv3.NewHuaBanApiV3(config.Name, config.Pass)
	huaBanApi.UploadOptions.ArgcFunc = func(filePath string) *huabanv3.Aigc {
		//todo 获取的逻辑代码
		return &huabanv3.Aigc{
			AigcCategory: "Stable Diffusion",
			Prompt:       "girl....",
		}
	}
	huaBanApi.UploadOptions.TagFunc = func(filePath string) []string {
		//todo 获取的逻辑代码
		return []string{"tag1", "tag2"}
	}
	
	huaBanApi.UploadOptions.TitleFunc = func(filePath string) string {
		//todo 获取的逻辑代码
		return "this is a title"
	}
}

```
可以结合[azure-vision](https://github.com/pzx521521/azurevision)的api来获取tag/title


### 使用代理 抓包或者换ip
```go
package main

import (
	"github.com/pzx521521/huabanv3"
	"net/http"
	"net/url"
)

func main() {

	proxyURL, _ := url.Parse("http://localhost:8888")
	// 创建一个带有代理的 Transport
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	// 创建一个带有自定义 Transport 的 Client
	client := &http.Client{
		Transport: transport,
	}
	huaBanApi := huabanv3.NewHuaBanApiV3(config.Name, config.Pass)
	huaBanApi.SetClient(client)
}
```



