package rx

import (
	"errors"
	"github.com/integration-system/go-cmp/cmp"
	"github.com/integration-system/isp-journal"
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/log"
	"github.com/integration-system/isp-journal/transfer"
	"github.com/integration-system/isp-lib/backend"
	"net"
)

var (
	ErrNotInitialize = errors.New("not initialized, call `ReceiveConfiguration` first")
)

type Config struct {
	log.Config
	Enable bool
}

type RxJournal struct {
	journal       journal.Journal
	serviceClient *backend.RxGrpcClient
	curState      state
}

func (j *RxJournal) ReceiveConfiguration(loggerConfig Config, moduleName string) {
	newState := state{
		Host:       getHost(),
		ModuleName: moduleName,
		Cfg:        loggerConfig,
	}
	if !cmp.Equal(j.curState, newState) {
		if j.journal != nil {
			_ = j.journal.Close()
			j.journal = nil
		}
		j.curState = newState

		if loggerConfig.Enable {
			j.journal = journal.NewFileJournal(
				loggerConfig.Config,
				moduleName,
				newState.Host,
				journal.WithAfterRotation(transfer.TransferAndDeleteLogFile(j.serviceClient, moduleName, newState.Host)),
			)
		}
	}
}

func (j *RxJournal) CollectAndTransferExistedLogs() {
	s := j.curState
	if !s.Cfg.Enable {
		return
	}

	logFiles, _ := log.CollectExistedLogs(s.Cfg.Config)
	if len(logFiles) > 0 {
		transfer.TransferAndDeleteLogFiles(j.serviceClient, s.ModuleName, s.Host)
	}
}

func (j *RxJournal) Log(level entry.Level, event string, req []byte, res []byte, err error) error {
	if j.journal == nil {
		return nil
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

type state struct {
	Host       string
	ModuleName string
	Cfg        Config
}

func NewDefaultRxJournal(journalServiceClient *backend.RxGrpcClient) *RxJournal {
	return &RxJournal{
		serviceClient: journalServiceClient,
	}
}

func getHost() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown"
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.To4().String()
}
