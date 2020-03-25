package rx

import (
	"errors"
	"github.com/integration-system/go-cmp/cmp"
	"github.com/integration-system/isp-journal"
	"github.com/integration-system/isp-journal/codes"
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/log"
	"github.com/integration-system/isp-journal/transfer"
	"github.com/integration-system/isp-lib/v2/backend"
	logger "github.com/integration-system/isp-log"
	"net"
)

const (
	RotationSignalEvent = "CONFIG:LOG_ROTATE"
)

var (
	ErrJournalClosed = errors.New("journal closed")
)

type Config struct {
	log.Config
	Enable               bool `schema:"Включение/отключение журналирования"`
	EnableRemoteTransfer bool `schema:"Отгрузка старых журналов,при включении старые файлы журналов будут отгружаются в сервис для isp-journal-service"`
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
			opts := make([]journal.Option, 0)
			if loggerConfig.EnableRemoteTransfer {
				opts = append(opts, journal.WithAfterRotation(transfer.TransferAndDeleteLogFile(j.serviceClient, moduleName, newState.Host)))
			}
			j.journal = journal.NewFileJournal(
				loggerConfig.Config,
				moduleName,
				newState.Host,
				opts...,
			)
		}
	}
}

func (j *RxJournal) CollectAndTransferExistedLogs() {
	s := j.curState
	if !s.Cfg.Enable || !s.Cfg.EnableRemoteTransfer {
		return
	}

	logFiles, _ := log.CollectExistedLogs(s.Cfg.Config)
	if len(logFiles) > 0 {
		transfer.TransferAndDeleteLogFiles(j.serviceClient, s.ModuleName, s.Host)(logFiles)
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

func (j *RxJournal) Rotate() error {
	if j.journal == nil {
		return ErrJournalClosed
	}
	return j.journal.Rotate()
}

func (j *RxJournal) Close() error {
	if j.journal == nil {
		return nil
	}
	err := j.journal.Close()
	j.journal = nil
	return err
}

func (j *RxJournal) SubscribeToRotationSignal() (event string, f func(_ []byte)) {
	return RotationSignalEvent, func(data []byte) {
		err := j.Rotate()
		if err != nil {
			logger.Warnf(codes.JournalingError, "log rotate signal: %v", err)
		}
	}
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
