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
		gzipReader, err := gzip.NewReader(pipe.Last())
		if err != nil {
			return nil, err
		}
		pipe.Unshift(gzipReader)
	}

	return &logReader{
		reader: pipe,
		filter: filter,
	}, nil
}

func (s *logReader) FilterNext() (*entry.Entry, error) {
	entry, err := entry.UnmarshalNext(s.reader)
	if err != nil {
		return nil, err
	}
	if s.filter.checkEntry(entry) {
		if ok, err := s.filter.checkTimeField(entry.Time); err != nil {
			return nil, err
		} else if !ok {
			return nil, nil
		}
		return entry, nil
	}
	return nil, nil
}

func (s *logReader) Close() error {
	return s.reader.Close()
}