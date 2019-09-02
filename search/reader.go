package search

import (
	"bufio"
	"compress/gzip"
	io2 "github.com/integration-system/isp-io"
	"github.com/integration-system/isp-journal/entry"
	"io"
)

type logReader struct {
	filter Filter
	reader io2.ReadPipe
}

func NewLogReader(reader io.Reader, gzipped bool, filter Filter) (*logReader, error) {
	pipe := io2.NewReadPipe(reader)

	pipe.Unshift(bufio.NewReaderSize(pipe.Last(), bufSize))
	if gzipped {
		if gzipReader, err := gzip.NewReader(pipe.Last()); err != nil {
			return nil, err
		} else {
			pipe.Unshift(gzipReader)
		}
	}

	return &logReader{
		reader: pipe,
		filter: filter,
	}, nil
}

func (s *logReader) FilterNext() (*entry.Entry, error) {
	if extractedEntry, err := entry.UnmarshalNext(s.reader); err != nil {
		return nil, err
	} else if s.filter.checkEntry(extractedEntry) {
		if ok, err := s.filter.checkTimeField(extractedEntry.Time); err != nil || !ok {
			return nil, err
		}
		return extractedEntry, nil
	}
	return nil, nil
}

func (s *logReader) Close() error {
	return s.reader.Close()
}
