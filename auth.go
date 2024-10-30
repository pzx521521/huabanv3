package huabanv3

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

func Login(account, password string) (cookie string, err error) {
	cookie, err = cookieLogin(account)
	if err != nil {
		log.Printf("尝试使用缓存Cookie登录失败，将尝试使用账密登录...%v", err)
		cookie, err = passwordLogin(account, password)
		if err != nil {
			return
		}
		return
	}
	return cookie, err
}
func passwordLogin(account, password string) (string, error) {
	// 使用账密登录
	log.Println("使用账密登录...")
	cookie, err := getCookie(account, password)
	if err != nil || cookie == "" {
		log.Printf("获取Cookie失败: %v", err)
		return cookie, err
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
func cookieLogin(account string) (cookie string, err error) {
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
	return cookie, nil
}
