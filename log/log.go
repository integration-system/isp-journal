package log

import (
	"fmt"
	io2 "github.com/integration-system/isp-io"
	"io"
	"os"
	"sync"
	"time"
)

type Logger interface {
	io.WriteCloser
	Rotate() error
}

type defaultLogger struct {
	c Config

	afterRotation func(prevFile LogFile)
	curSize       int64
	wrLock        sync.Mutex
	curWr         io2.WritePipe
	rotateChan    chan struct{}
	rotateErrChan chan error
	closeChan     chan struct{}
}

func (l *defaultLogger) Write(p []byte) (int, error) {
	l.wrLock.Lock()
	defer l.wrLock.Unlock()

	writeLen := int64(len(p))
	if writeLen > l.c.GetMaxSizeInBytes() {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum file size %d", writeLen, l.c.GetMaxSizeInBytes(),
		)
	}

	if l.curWr == nil {
		if err := l.openExistedOrNew(len(p)); err != nil {
			return 0, err
		}
	}

	if l.curSize+writeLen > l.c.GetMaxSizeInBytes() {
		if err := l.rotateWithoutLock(); err != nil {
			return 0, err
		}
	}

	n, err := l.curWr.Write(p)
	l.curSize += int64(n)

	return n, err
}

func (l *defaultLogger) Close() error {
	l.wrLock.Lock()
	defer l.wrLock.Unlock()

	close(l.closeChan)
	close(l.rotateChan)
	close(l.rotateErrChan)

	if l.curWr != nil {
		return l.curWr.Close()
	}

	return nil
}

func (l *defaultLogger) Rotate() error {
	l.wrLock.Lock()
	defer l.wrLock.Unlock()

	return l.rotateWithoutLock()
}

func (l *defaultLogger) rotateWithoutLock() error {
	select {
	case l.rotateChan <- struct{}{}:
	default:

	}
	return <-l.rotateErrChan
}

func (l *defaultLogger) prepare() {
	if l.c.GetRotateTimeout() != 0 {
		go func() {
			for {
				select {
				case <-time.After(l.c.GetRotateTimeout()):
					_ = l.Rotate()
				case <-l.closeChan:
					return
				}
			}
		}()
	}

	go l.awaitRotationSignal()
}

func (l *defaultLogger) awaitRotationSignal() {
	for range l.rotateChan {
		if l.curWr != nil {
			if err := l.curWr.Close(); err != nil {
				l.rotateErrChan <- err
				return
			}
		}

		if pipe, oldFile, err := openNewAndRenameExisted(l.c); err != nil {
			l.rotateErrChan <- err
		} else {
			l.curWr = pipe
			l.curSize = 0
			if l.afterRotation != nil && oldFile != "" {
				go func() {
					if log, err := MakeLogFile(l.c, oldFile); err == nil {
						l.afterRotation(*log)
					}
				}()
			}
			l.rotateErrChan <- nil
		}
	}
}

// openExistedOrNew opens the logfile if it exists and if the current write
// would not put it over MaxSize.  If there is no such file or the write would
// put it over the MaxSize, a new file is created.
func (l *defaultLogger) openExistedOrNew(writeLen int) error {
	filename := l.c.GetFilename()
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		if p, _, err := openNewAndRenameExisted(l.c); err != nil {
			return err
		} else {
			l.curWr = p
			l.curSize = 0
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	if info.Size()+int64(writeLen) >= l.c.GetMaxSizeInBytes() {
		return l.rotateWithoutLock()
	}

	p, err := makePipe(l.c, filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		p, _, err = openNewAndRenameExisted(l.c)
	}
	if err != nil {
		return err
	}
	l.curWr = p
	l.curSize = info.Size()
	return nil
}

func NewDefaultLogger(config Config, opts ...Option) Logger {
	l := &defaultLogger{
		c:             config,
		rotateChan:    make(chan struct{}),
		rotateErrChan: make(chan error),
		closeChan:     make(chan struct{}),
	}

	for _, opt := range opts {
		opt(l)
	}

	l.prepare()

	return l
}
