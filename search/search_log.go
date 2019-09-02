package search

import (
	"github.com/integration-system/isp-journal/entry"
)

type searchLog struct {
	entriesHandler func(*entry.Entry) (bool, error)
	s              *SyncSearchLog
	baseDir        string
}

func NewSearchLog(entriesHandler func(*entry.Entry) (continueRead bool, err error), baseDir string) *searchLog {
	return &searchLog{
		entriesHandler: entriesHandler,
		baseDir:        baseDir,
	}
}

func (s *searchLog) Search(req SearchRequest) error {
	var err error
	if s.s, err = NewSyncSearchService(req, s.baseDir); err != nil {
		return err
	} else if err := s.extractData(); err != nil {
		return err
	}
	return nil
}

func (s *searchLog) extractData() error {
	for {
		if extractedEntry, hasMore, err := s.s.Next(); err != nil || !hasMore {
			return err
		} else if extractedEntry != nil {
			if continueRead, err := s.entriesHandler(extractedEntry); err != nil {
				return err
			} else if !continueRead {
				return nil
			}
		}
	}
}
