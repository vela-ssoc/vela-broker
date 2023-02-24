package reqresp

import "time"

type Syscmd struct {
	Timeout time.Duration `json:"timeout" query:"timeout"`
	Cmd     string        `json:"cmd"     query:"cmd"     validate:"required,lte=30"`
	Args    []string      `json:"args"    query:"args"    validate:"lte=100"`
}
