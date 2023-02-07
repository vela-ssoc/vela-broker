package httpclient

import "fmt"

type Error struct {
	Code int
	Text []byte
}

func (e *Error) Error() string {
	return fmt.Sprintf("http response status %d, message is: %s", e.Code, e.Text)
}

type Description struct {
	Filename string `json:"filename"`
	Checksum string `json:"checksum"`
	Filesize int64  `json:"filesize"`
}
