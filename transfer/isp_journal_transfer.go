package transfer

import (
	"github.com/integration-system/isp-journal/log"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/streaming"
	"google.golang.org/grpc/metadata"

	"os"
)

const (
	transferMethod = "isp-journal/log/transfer"
)

func TransferAndDeleteLogFile(client *backend.RxGrpcClient) func(file log.LogFile) {
	return func(log log.LogFile) {
		doTransfer(client, log)
	}
}

func TransferAndDeleteLogFiles(client *backend.RxGrpcClient) func(logs []log.LogFile) {
	return func(logs []log.LogFile) {
		for _, f := range logs {
			doTransfer(client, f)
		}
	}
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

func statToFileHeader(f os.FileInfo) streaming.BeginFile {
	return streaming.BeginFile{
		FileName:      f.Name(),
		FormDataName:  f.Name(),
		ContentType:   "application/binary",
		ContentLength: f.Size(),
	}
}
