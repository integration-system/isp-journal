package entry

import "time"

type Level string

const (
	LevelError = "ERROR"
	LevelInfo  = "OK"
	LevelWarn  = "WARN"
)

const (
	timeFormat = "2006-01-02T15:04:05.999-07:00"
)

func FormatTime(time time.Time) string {
	return time.Format(timeFormat)
}

func ParserTime(s string) (time.Time, error) {
	return time.Parse(timeFormat, s)
}
