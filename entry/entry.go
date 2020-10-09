package entry

import (
	"time"
)

const (
	LevelError = "ERROR"
	LevelInfo  = "OK"
	LevelWarn  = "WARN"

	timeFormat = "2006-01-02T15:04:05.999-07:00"
)

type Level string

type Entry struct {
	ModuleName string `json:"moduleName,omitempty"`
	Host       string `json:"host,omitempty"`
	Event      string `json:"event,omitempty"`
	Level      string `json:"level,omitempty"`
	Time       string `json:"time,omitempty"`
	Request    []byte `json:"request,omitempty"`
	Response   []byte `json:"response,omitempty"`
	ErrorText  string `json:"errorText,omitempty"`
}

func ParserTime(s string) (time.Time, error) {
	return time.Parse(timeFormat, s)
}

func FormatTime(time time.Time) string {
	return time.Format(timeFormat)
}
