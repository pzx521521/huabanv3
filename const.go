package huabanv3

import "net/http"

const COOKIEJSONPATH = "./cookie.json"
const DOMAIN = "https://huaban.com"
const UA = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36"

func SetHeader(req *http.Request, headers map[string]string) {
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
		Color int `json:"color"`
		Ratio int `json:"ratio"`
	} `json:"colors"`
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
type Aigc struct {
	AigcCategory string `json:"aigc_category,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
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
