package log

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	io2 "github.com/integration-system/isp-io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	backupTimeFormat = "2006-01-02T15-04-05.000"
)

type LogFile struct {
	os.FileInfo
	CreatedAt  time.Time
	FullPath   string
	Compressed bool
}

func CollectExistedLogs(loggerConfig Config) ([]LogFile, error) {
	files, err := ioutil.ReadDir(loggerConfig.GetDirectory())
	if err != nil {
		return nil, err
	}
	logFiles := make([]LogFile, 0)

	prefix, ext := loggerConfig.GetFilePrefixAndExt()

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if t, err := parseTimeFormFilename(f.Name(), prefix, ext); err == nil {
			logFiles = append(logFiles, LogFile{
				Compressed: path.Ext(f.Name()) == ".gz",
				FileInfo:   f,
				CreatedAt:  t,
				FullPath:   path.Join(loggerConfig.GetDirectory(), f.Name()),
			})
		}
		// error parsing means that the suffix at the end was not generated
		// by lumberjack, and therefore it's not a backup file.
	}

	return logFiles, nil
}

func MakeLogFile(c Config, filepath string) (*LogFile, error) {
	if info, err := os.Stat(filepath); err != nil {
		return nil, err
	} else {
		prefix, ext := c.GetFilePrefixAndExt()
		if t, err := parseTimeFormFilename(info.Name(), prefix, ext); err != nil {
			return nil, err
		} else {
			return &LogFile{
				FullPath:   filepath,
				CreatedAt:  t,
				FileInfo:   info,
				Compressed: c.IsCompress(),
			}, nil
		}
	}
}

func parseTimeFormFilename(filename, prefix, ext string) (time.Time, error) {
	if !strings.HasPrefix(filename, prefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}
	if !strings.HasSuffix(filename, ext) {
		return time.Time{}, errors.New("mismatched extension")
	}
	ts := filename[len(prefix) : len(filename)-len(ext)]
	return time.Parse(backupTimeFormat, ts)
}

// openNewAndRenameExisted opens a new log file for writing, moving any old log file out of the
// way.  This methods assumes the file has already been closed.
func openNewAndRenameExisted(c Config) (io2.WritePipe, string, error) {
	err := os.MkdirAll(c.GetDirectory(), 0755)
	if err != nil {
		return nil, "", fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	name := c.GetFilename()
	mode := os.FileMode(0600)
	info, err := os.Stat(name)
	newname := ""
	if err == nil {
		// Copy the mode off the old logfile.
		mode = info.Mode()
		// move the existing file
		newname = getBackupFileName(name)
		if err := os.Rename(name, newname); err != nil {
			return nil, "", fmt.Errorf("can't rename log file: %s", err)
		}

		/*// this is a no-op anywhere but linux
		if err := chown(name, info); err != nil {
			return err
		}*/
	}

	pipe, err := makePipe(c, name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return nil, "", err
	}

	return pipe, newname, nil
}

func makePipe(c Config, srcFile string, flag int, mode os.FileMode) (io2.WritePipe, error) {
	f, err := os.OpenFile(srcFile, flag, mode)
	if err != nil {
		return nil, err
	}
	p := io2.NewWritePipe(f)

	if c.IsBuffered() {
		bufWr := bufio.NewWriterSize(p.Last(), c.GetBufferSize())
		p.Unshift(bufWr)
	}

	if c.IsCompress() {
		gzipWr := gzip.NewWriter(p.Last())
		p.Unshift(gzipWr)
	}

	return p, nil
}

func getBackupFileName(name string) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]

	timestamp := time.Now().UTC().Format(backupTimeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}
