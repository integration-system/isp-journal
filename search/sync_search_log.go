package search

import (
	"fmt"
	"github.com/integration-system/isp-journal/entry"
	"io"
	"os"
)

type SyncSearchLog struct {
	filter        Filter
	files         []string
	currentReader *logReader
}

// return next matched log entry, never return io.EOF, if data source exhausted return false in second param
// idempotent return the same error while reading or opening file
func (s *SyncSearchLog) Next() (*entry.Entry, bool, error) {
	if s.currentReader == nil {
		hasMore, err := s.openNextReader()
		if err != nil {
			return nil, false, err
		}
		if !hasMore {
			return nil, false, nil
		}
	}

	for {
		entry, err := s.currentReader.FilterNext()
		if err != nil {
			if err == io.EOF {
				_ = s.currentReader.Close()
				hasMore, err := s.openNextReader()
				if err != nil {
					return nil, false, err
				}
				if !hasMore {
					return nil, false, nil
				}
			}
			return nil, false, err
		}
		if entry != nil {
			return entry, true, nil
		}
	}
}

func (s *SyncSearchLog) openNextReader() (bool, error) {
	if len(s.files) == 0 {
		return false, nil
	}
	currentFile, files := s.files[0], s.files[1:]

	file, err := os.Open(currentFile)
	if err != nil {
		return false, fmt.Errorf("could not open file %s: %v", currentFile, err)
	}
	currentReader, err := NewLogReader(file, true, s.filter)
	if err != nil {
		return false, fmt.Errorf("could not open log reader %s: %v", currentFile, err)
	}

	s.files = files
	s.currentReader = currentReader

	return true, err
}

func NewSyncSearch(req SearchRequest, baseDir string) (*SyncSearchLog, error) {
	filter, err := NewFilter(req)
	if err != nil {
		return nil, err
	}

	files, err := findAllMatchedFiles(filter, baseDir)
	if err != nil {
		return nil, err
	}

	return &SyncSearchLog{
		filter: filter,
		files:  files,
		closed: false,
	}, nil
}
