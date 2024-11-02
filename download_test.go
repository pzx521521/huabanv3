package huabanv3

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"
)

func TestDownloadBoard(t *testing.T) {
	err := DownloadBoard(94004345, 2, false, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestGetImgInfos(t *testing.T) {
	infos, err := GetImgInfos(http.DefaultClient, 94004345, nil)
	if err != nil {
		return
	}
	var result []string
	for _, info := range infos {
		imgUrl := fmt.Sprintf("https://%s.huaban.com/%s",
			info.File.Bucket, info.File.Key)
		result = append(result, imgUrl)
	}
	j, _ := json.Marshal(result)
	log.Printf("%s\n", j)
}
