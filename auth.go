package huabanv3

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
)

func Login(client *http.Client, account, password string, header map[string]string) (name string, err error) {
	name, err = cookieLogin(client, account, header)
	if err != nil {
		log.Printf("使用缓存cookie获取用户名失败,刷新cookie..%v\n", err)
		cookie, err := passwordLogin(client, account, password)
		if err != nil {
			return "", err
		}
		header["Cookie"] = cookie
		name, err = getUserName(http.DefaultClient, header)
		if err != nil {
			return "", err
		}
		return name, nil
	}
	return name, nil
}
func passwordLogin(client *http.Client, account, password string) (string, error) {
	// 使用账密登录
	log.Println("使用账密登录...")
	cookie, err := getCookie(client, account, password)
	if err != nil || cookie == "" {
		log.Printf("获取Cookie失败: %v", err)
		return "", err
	}
	log.Printf("账密登录成功，获取到cookies:%s", cookie)
	saveCookie(account, cookie)
	return cookie, nil
}

func saveCookie(account, cookie string) {
	content, err := os.ReadFile(COOKIEJSONPATH)
	cookieMap := map[string]string{}
	if !os.IsNotExist(err) {
		json.Unmarshal(content, &cookieMap)
	}
	cookieMap[account] = cookie
	newContent, _ := json.Marshal(cookieMap)
	if err := os.WriteFile(COOKIEJSONPATH, newContent, 0644); err == nil {
		log.Println("缓存Cookie成功")
	}
}
func cookieLogin(client *http.Client, account string, header map[string]string) (name string, err error) {
	// 检查 cookie.json 文件是否存在且有效
	info, err := os.Stat(COOKIEJSONPATH)
	if err != nil || info.IsDir() {
		return
	}

	content, err := os.ReadFile(COOKIEJSONPATH)
	if err != nil {
		return
	}

	var cookieMap map[string]string
	err = json.Unmarshal(content, &cookieMap)
	if err != nil {
		return
	}
	cookie, ok := cookieMap[account]
	if !ok {
		err = errors.New("找到不发送缓存Cookie对应的用户名...")
		return
	}
	header["Cookie"] = cookie
	log.Printf("尝试使用缓存Cookie获取用户名...")
	name, err = getUserName(client, header)
	if err != nil {
		return
	}
	return name, nil
}
