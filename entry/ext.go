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

func ParserTime(s string) (time.Time, error) {
	return time.Parse(timeFormat, s)
}

func FormatTime(time time.Time) string {
	return time.Format(timeFormat)
}
