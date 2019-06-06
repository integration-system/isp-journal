package rx

import (
	"errors"
	"github.com/integration-system/go-cmp/cmp"
	"github.com/integration-system/isp-journal"
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/log"
	"github.com/integration-system/isp-journal/transfer"
	"github.com/integration-system/isp-lib/backend"
)

var (
	ErrNotInitialize = errors.New("not initialized, call `ReceiveConfiguration` first")
)

type RxJournal struct {
	journal       journal.Journal
	serviceClient *backend.RxGrpcClient
	lastConfig    map[string]interface{}
}

func (j *RxJournal) ReceiveConfiguration(loggerConfig log.Config, moduleName, host string) {
	newConfig := map[string]interface{}{
		"moduleName": moduleName,
		"host":       host,
		"config":     loggerConfig,
	}
	if !cmp.Equal(j.lastConfig, newConfig) {
		_ = j.journal.Close()
		j.journal = journal.NewFileJournal(
			loggerConfig,
			moduleName,
			host,
			journal.WithAfterRotation(transfer.TransferAndDeleteLogFile(j.serviceClient, moduleName, host)),
			journal.WithCollectingExistedLogs(transfer.TransferAndDeleteLogFiles(j.serviceClient, moduleName, host)),
		)
		j.lastConfig = newConfig
	}
}

func (j *RxJournal) Log(level entry.Level, event string, req []byte, res []byte, err error) error {
	if j.journal == nil {
		return ErrNotInitialize
	}
	return j.journal.Log(level, event, req, res, err)
}

func (j *RxJournal) Info(event string, req []byte, res []byte) error {
	return j.Log(entry.LevelInfo, event, req, res, nil)
}

func (j *RxJournal) Warn(event string, req []byte, res []byte, err error) error {
	return j.Log(entry.LevelWarn, event, req, res, err)
}

func (j *RxJournal) Error(event string, req []byte, res []byte, err error) error {
	return j.Log(entry.LevelError, event, req, res, err)
}

func (j *RxJournal) Close() error {
	return j.journal.Close()
}

func NewDefaultRxJournal(journalServiceClient *backend.RxGrpcClient) *RxJournal {
	return &RxJournal{
		serviceClient: journalServiceClient,
		lastConfig:    make(map[string]interface{}),
	}
}