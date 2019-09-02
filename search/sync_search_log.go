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

func NewSyncSearchService(req SearchRequest, baseDir string) (*SyncSearchLog, error) {
	if filter, err := NewFilter(req); err != nil {
		return nil, err
	} else if files, err := findAllMatchedFiles(filter, baseDir); err != nil {
		return nil, err
	} else {
		return &SyncSearchLog{
			filter: filter,
			files:  files,
		}, nil
	}
}

// return next matched log entry, never return io.EOF, if data source exhausted return false in second param
// idempotent return the same error while reading or opening file
func (s *SyncSearchLog) Next() (*entry.Entry, bool, error) {
	if s.currentReader == nil {
		if hasMore, err := s.openNextReader(); err != nil || !hasMore {
			return nil, false, err
		}
	}

	for {
		if entry, err := s.currentReader.FilterNext(); err != nil {
			if err == io.EOF {
				_ = s.currentReader.Close()
				s.currentReader = nil
				if hasMore, err := s.openNextReader(); err != nil || !hasMore {
					return nil, false, err
				} else {
					continue
				}
			}
			return nil, false, err
		} else if entry != nil {
			return entry, true, nil
		}
	}
}

func (s *SyncSearchLog) openNextReader() (bool, error) {
	for {
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
			if err == io.EOF {
				s.files = files
				continue
			}
			return false, fmt.Errorf("could not open log reader %s: %v", currentFile, err)
		}
		s.files = files
		s.currentReader = currentReader
		return true, err
	}
}
