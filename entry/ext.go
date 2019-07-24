package entry

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"sync"
	"time"
)

type Level string

const (
	LevelError = "ERROR"
	LevelInfo  = "OK"
	LevelWarn  = "WARN"
)

const (
	timeFormat = "2006-01-02T15:04:05.999-07:00"
)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4096))
		},
	}
)

func FormatTime(time time.Time) string {
	return time.Format(timeFormat)
}

func ParserTime(s string) (time.Time, error) {
	return time.Parse(timeFormat, s)
}

func MarshalToBytes(entry *Entry) ([]byte, error) {
	bytes, err := proto.Marshal(entry)
	if err != nil {
		return nil, err
	}
	l := make([]byte, 4)
	binary.LittleEndian.PutUint32(l, uint32(len(bytes)))
	return append(l, bytes...), nil
}

func UnmarshalNext(r io.Reader) (*Entry, error) {
	buf := pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		pool.Put(buf)
	}()

	n, err := io.CopyN(buf, r, 4)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, errors.New("expecting int32 data length prefix")
	}
	l := int64(binary.LittleEndian.Uint32(buf.Bytes()))

	buf.Reset()
	n, err = io.CopyN(buf, r, l)
	if err != nil {
		return nil, err
	}
	if n != l {
		return nil, fmt.Errorf("not enough %d bytes", l-n)
	}

	e := &Entry{}
	if err := proto.Unmarshal(buf.Bytes(), e); err != nil {
		return nil, err
	}

	return e, nil
}
