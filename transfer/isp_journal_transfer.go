package transfer

import (
	"fmt"
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/log"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/streaming"
	"google.golang.org/grpc/metadata"
	"time"

	"os"
)

const (
	transferMethod = "journal/log/transfer"

	moduleNameField = "moduleName"
	createdAtField  = "createdAt"
	hostField       = "host"
)

type LogInfo struct {
	ModuleName string
	Host       string
	CreatedAt  time.Time
}

func TransferAndDeleteLogFile(client *backend.RxGrpcClient) func(file log.LogFile) {
	return func(log log.LogFile) {
		if log.Size() > 0 {
			doTransfer(client, log)
		}
	}
}

func TransferAndDeleteLogFiles(client *backend.RxGrpcClient) func(logs []log.LogFile) {
	return func(logs []log.LogFile) {
		for _, f := range logs {
			if f.Size() > 0 {
				doTransfer(client, f)
			}
		}
	}
}

func GetLogInfo(bf streaming.BeginFile) (*LogInfo, error) {
	moduleName, err := bf.FormData.GetStringValue(moduleNameField)
	if err != nil {
		return nil, err
	}

	host, err := bf.FormData.GetStringValue(hostField)
	if err != nil {
		return nil, err
	}

	createdAt, err := bf.FormData.GetStringValue(createdAtField)
	if err != nil {
		return nil, err
	}

	createdAtTime, err := entry.ParserTime(createdAt)
	if err != nil {
		return nil, fmt.Errorf("invalid '%s' time format: %v", createdAtField, err)
	}

	return &LogInfo{
		ModuleName: moduleName,
		Host:       host,
		CreatedAt:  createdAtTime,
	}, nil
}

func doTransfer(client *backend.RxGrpcClient, f log.LogFile) {
	err := client.Visit(func(c *backend.InternalGrpcClient) error {
		return c.InvokeStream(transferMethod, -1, func(stream streaming.DuplexMessageStream, md metadata.MD) error {
			return streaming.WriteFile(stream, f.FullPath, statToFileHeader(f))
		})
	})

	if err != nil {
		logger.Errorf("could not transfer log file '%s': %v", f.FullPath, err)
	} else if err := os.Remove(f.FullPath); err != nil {
		logger.Warnf("could not remove log file '%s': %v", f.FullPath, err)
	} else {
		logger.Debugf("log '%s' successfully transferred", f.FullPath)
	}
}

func statToFileHeader(f log.LogFile, moduleName, host string) streaming.BeginFile {
	formData := map[string]interface{}{
		moduleNameField: moduleName,
		hostField:       host,
		createdAtField:  entry.FormatTime(f.CreatedAt),
	}
	return streaming.BeginFile{
		FileName:      f.Name(),
		FormDataName:  f.Name(),
		ContentType:   "application/binary",
		ContentLength: f.Size(),
		FormData:      formData,
	}
}
