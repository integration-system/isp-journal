package journal

import (
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/log"
	"io"
	"time"
)

type Journal interface {
	io.Closer
	Log(status entry.Level, event string, req []byte, res []byte, err error) error
	Info(event string, req []byte, res []byte) error
	Warn(event string, req []byte, res []byte, err error) error
	Error(event string, req []byte, res []byte, err error) error
}

type fileJournal struct {
	log log.Logger

	host                 string
	moduleName           string
	afterRotation        func(log log.LogFile)
	existedLogsCollector func(logs []log.LogFile)
}

func (j *fileJournal) Log(level entry.Level, event string, req []byte, res []byte, err error) error {
	e := &entry.Entry{
		ModuleName: j.moduleName,
		Host:       j.host,
		Event:      event,
		Time:       entry.FormatTime(time.Now().UTC()),
		Level:      string(level),
		Request:    req,
		Response:   res,
	}
	if err != nil {
		e.ErrorText = err.Error()
	}

	bytes, err := entry.MarshalToBytes(e)
	if err != nil {
		return err
	}

	_, err = j.log.Write(bytes)
	return err
}

func (j *fileJournal) Info(event string, req []byte, res []byte) error {
	return j.Log(entry.LevelInfo, event, req, res, nil)
}

func (j *fileJournal) Warn(event string, req []byte, res []byte, err error) error {
	return j.Log(entry.LevelWarn, event, req, res, err)
}

func (j *fileJournal) Error(event string, req []byte, res []byte, err error) error {
	return j.Log(entry.LevelError, event, req, res, err)
}

func (j *fileJournal) Close() error {
	return j.log.Close()
}

func NewFileJournal(loggerConfig log.Config, moduleName, host string, opts ...Option) Journal {
	j := &fileJournal{
		moduleName: moduleName,
		host:       host,
	}

	for _, opt := range opts {
		opt(j)
	}

	j.log = log.NewDefaultLogger(loggerConfig, log.WithAfterRotation(j.afterRotation))

	return j
}
