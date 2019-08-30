package search

import (
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-lib/logger"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

const (
	dirLayout  = "2006-01-02"
	fileLayout = "2006-01-02T15-04-05.000"

	bufSize   = 64 * 1024
	fileSplit = "__"
	fileEnd   = ".log"
)

type searchLog struct {
	entriesHandler func(*entry.Entry) (bool, error)
	filter         Filter
	baseDir        string
}

func NewSearchLog(entriesHandler func(*entry.Entry) (continueRead bool, err error), baseDir string) *searchLog {
	return &searchLog{
		entriesHandler: entriesHandler,
		baseDir:        baseDir,
	}
}

func (s *searchLog) Search(req SearchRequest) error {
	filter, err := NewFilter(req)
	if err != nil {
		return err
	}
	s.filter = filter

	if arrayOfPath, err := s.getFilesPath(req.ModuleName); err != nil {
		return err
	} else if len(arrayOfPath) > 0 {
		if err := s.readFiles(arrayOfPath); err != nil {
			return err
		}
	}
	return nil
}

func (s *searchLog) readFiles(files []string) error {
	for _, filePath := range files {
		if continueRead, err := s.extractData(filePath); err != nil {
			return err
		} else if !continueRead {
			return nil
		}
	}
	return nil
}

func (s *searchLog) extractData(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	reader, err := NewLogReader(file, true, s.filter)
	if reader != nil {
		defer func() { _ = reader.Close() }()
	}
	if err != nil {
		if err == io.EOF {
			return true, nil
		}
		return false, err
	}

	for {
		if extractedEntry, err := reader.FilterNext(); err != nil {
			if err == io.EOF {
				return true, nil
			}
			return false, err
		} else if extractedEntry != nil {
			if continueRead, err := s.entriesHandler(extractedEntry); err != nil {
				return false, err
			} else if !continueRead {
				return false, nil
			}
		}
	}
}
