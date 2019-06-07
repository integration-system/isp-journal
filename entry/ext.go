package entry

import (
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
			return make([]byte, 4096)
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
	buf := pool.Get().([]byte)

	uintBuf, dataBuf := buf[:4], buf[4:]
	n, err := r.Read(uintBuf)
	if err != nil {
		pool.Put(buf)
		return nil, err
	}
	if n != 4 {
		pool.Put(buf)
		return nil, errors.New("expecting int32 data length prefix")
	}
	l := int(binary.LittleEndian.Uint32(uintBuf))

	size := len(dataBuf)
	toRead := dataBuf
	pooled := buf
	if size < l {
		toRead = append(dataBuf, make([]byte, l-size)...)
		pooled = toRead
	} else if size > l {
		toRead = dataBuf[:l]
	}

	n, err = r.Read(toRead)
	if err != nil {
		pool.Put(pooled)
		return nil, err
	}
	if n != l {
		pool.Put(pooled)
		return nil, fmt.Errorf("not enough %d bytes", l-n)
	}

	e := &Entry{}
	if err := proto.Unmarshal(toRead, e); err != nil {
		pool.Put(pooled)
		return nil, err
	}

	pool.Put(pooled)

	return e, nil
}
