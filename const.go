package huabanv3

import (
	"fmt"
	"net/http"
)

type UploadIgnoreError struct {
	Msg string `json:"msg"`
	Err int    `json:"err"`
}

func (e *UploadIgnoreError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Err, e.Msg)
}

func (e *UploadIgnoreError) Is(target error) bool {
	// 比较错误代码 500 是滑动验证, 此时应停止上传
	return e.Err != 500
}

//

const COOKIEJSONPATH = "./cookie.json"
const DOMAIN = "https://huaban.com"
const UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"

func SetHeader(req *http.Request, headers map[string]string) {
	if headers == nil {
		return
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func SetHeaderUA(req *http.Request) {
	req.Header.Set("User-Agent", UA)
}

func SetHeaderAsJson(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}

type User struct {
	UserId    int    `json:"user_id"`
	Username  string `json:"username"`
	Urlname   string `json:"urlname"`
	CreatedAt int    `json:"created_at"`
}

// Board 表示看板信息
type Board struct {
	BoardId     int    `json:"board_id"`
	UserId      int    `json:"user_id"`
	LikeCount   int    `json:"like_count"`
	FollowCount int    `json:"follow_count"`
	IsPrivate   int    `json:"is_private"`
	Title       string `json:"title"`
	PinCount    int    `json:"pin_count"`
	Seq         int    `json:"seq"`
}

type File struct {
	Id     int    `json:"id"`
	Farm   string `json:"farm"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Type   string `json:"type"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Frames int    `json:"frames"`
	Theme  string `json:"theme"`
	UserId int    `json:"user_id"`
	Colors []struct {
		Color int     `json:"color"`
		Ratio float64 `json:"ratio"`
	} `json:"colors"`
}

const (
	ImgSize_SQ180  = "_sq180webp"
	ImgSize_SQ490  = "_sq490webp"
	ImgSize_W240   = "_fw240webp"
	ImgSize_W658   = "_fw658webp"
	ImgSize_W1200  = "_fw1200webp"
	ImgSize_Origin = ""
)

func (f *File) GetImgUrl(suffix string) string {
	return fmt.Sprintf("https://%s.huaban.com/%s%s",
		f.Bucket, f.Key, suffix)
}

type BatchInfo struct {
	BoardId int          `json:"board_id"`
	Pins    []*UploadPin `json:"pins"`
}
type UploadPin struct {
	FileId int      `json:"file_id"`
	Text   string   `json:"text"` //文件名
	Tags   []string `json:"tags,omitempty"`
	Aigc   *Aigc    `json:"aigc,omitempty"`
}
type PutPinInfo struct {
	PinId   int64    `json:"pin_id"`
	BoardId int      `json:"board_id"`
	Text    string   `json:"text,omitempty"`
	Link    string   `json:"link,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Aigc    *Aigc    `json:"aigc,omitempty"`
}
type Aigc struct {
	AigcCategory string `json:"aigc_category,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Model        string `json:"model,omitempty"`
}

// Pin 表示图钉信息
type Pin struct {
	Tags      []string    `json:"tags"`
	PinID     int64       `json:"pin_id"`
	LikeCount int         `json:"like_count"`
	File      File        `json:"file"`
	BoardId   int         `json:"board_id"`
	FileId    int         `json:"file_id"`
	RawText   string      `json:"raw_text"`
	Extra     interface{} `json:"extra"`
}

// BoardsResponse 表示响应结构
type PinsResponse struct {
	Pins  []*Pin `json:"pins"`
	Board *Board `json:"board"`
}

// BoardsResponse 表示响应结构
type BoardsResponse struct {
	User   User     `json:"user"`
	Boards []*Board `json:"boards"`
}

type BoardResponse struct {
	User  User   `json:"user"`
	Board *Board `json:"board"`
}

func (r *BoardsResponse) ExistBoard(boardName string) *Board {
	for _, board := range r.Boards {
		if board.Title == boardName {
			return board
		}
	}
	return nil
}
