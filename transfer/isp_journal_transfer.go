package transfer

import (
	"fmt"
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/log"
	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/streaming"
	logger "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	"google.golang.org/grpc/metadata"
	"os"
	"time"
)

const (
	transferMethod = "journal/log/transfer"

	moduleNameField = "moduleName"
	createdAtField  = "createdAt"
	hostField       = "host"

	gzipContent   = "application/gzip"
	binaryContent = "application/binary"
)

type LogInfo struct {
	ModuleName string
	Host       string
	Compressed bool
	CreatedAt  time.Time
}

func TransferAndDeleteLogFile(client *backend.RxGrpcClient, moduleName, host string) func(file log.LogFile) {
	return func(log log.LogFile) {
		doTransfer(client, log, moduleName, host)
	}
}

func TransferAndDeleteLogFiles(client *backend.RxGrpcClient, moduleName, host string) func(logs []log.LogFile) {
	return func(logs []log.LogFile) {
		for _, f := range logs {
			doTransfer(client, f, moduleName, host)
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

	compressed := false
	if bf.ContentType == gzipContent {
		compressed = true
	}

	return &LogInfo{
		ModuleName: moduleName,
		Host:       host,
		Compressed: compressed,
		CreatedAt:  createdAtTime,
	}, nil
}

func doTransfer(client *backend.RxGrpcClient, f log.LogFile, moduleName, host string) {
	err := client.Visit(func(c *backend.InternalGrpcClient) error {
		return c.InvokeStream(transferMethod, -1, func(stream streaming.DuplexMessageStream, md metadata.MD) error {
			return streaming.WriteFile(stream, f.FullPath, statToFileHeader(f, moduleName, host))
		})
	})

	if err != nil {
		logger.Errorf(stdcodes.ModuleDefaultRCReadError, "could not transfer log file '%s': %v", f.FullPath, err)
	} else if err := os.Remove(f.FullPath); err != nil {
		logger.Warnf(stdcodes.ModuleDefaultRCReadError, "could not remove log file '%s': %v", f.FullPath, err)
	} else {
		logger.Debugf(stdcodes.ModuleDefaultRCReadError, "log '%s' successfully transferred", f.FullPath)
	}
}

func statToFileHeader(f log.LogFile, moduleName, host string) streaming.BeginFile {
	formData := map[string]interface{}{
		moduleNameField: moduleName,
		hostField:       host,
		createdAtField:  entry.FormatTime(f.CreatedAt),
	}
	ct := binaryContent
	if f.Compressed {
		ct = gzipContent
	}
	return streaming.BeginFile{
		FileName:      f.Name(),
		FormDataName:  f.Name(),
		ContentType:   ct,
		ContentLength: f.Size(),
		FormData:      formData,
	}
}
