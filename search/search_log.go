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

func (s *searchLog) getFilesPath(moduleName string) ([]string, error) {
	if dirs, err := s.findDirs(); err != nil {
		return nil, err
	} else if len(dirs) == 0 {
		return nil, nil
	} else {
		return s.findFiles(dirs, moduleName)
	}
}

func (s *searchLog) findDirs() ([]string, error) {
	f := time.Date(
		s.filter.from.Year(), s.filter.from.Month(), s.filter.from.Day(),
		0, 0, 0, 0, s.filter.from.Location())
	t := time.Date(
		s.filter.to.Year(), s.filter.to.Month(), s.filter.to.Day(),
		0, 0, 0, 0, s.filter.to.Location()).AddDate(0, 0, 1)
	dirs := make([]string, 0)
	if arrayDateDir, err := ioutil.ReadDir(s.baseDir); err != nil {
		return nil, err
	} else {
		for _, dateDirInfo := range arrayDateDir {
			if dirName, err := time.Parse(dirLayout, dateDirInfo.Name()); err != nil {
				continue
			} else if dirName.After(f) && dirName.Before(t) {
				dirs = append(dirs, dateDirInfo.Name())
			}
		}
	}
	return dirs, nil
}

func (s *searchLog) findFiles(dirs []string, middleFile string) ([]string, error) {
	response := make([]string, 0)
	for _, dir := range dirs {
		dir := path.Join(s.baseDir, dir, middleFile)
		filesInfo, err := ioutil.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return nil, err
			}
		}
		for _, fileInfo := range filesInfo {
			fileName := strings.Split(fileInfo.Name(), fileSplit)
			if !s.filter.checkHost(fileName[0]) {
				continue
			}
			if len(fileName) < 2 {
				logger.Warnf("invalid file name '%s'", fileInfo.Name())
				continue
			}
			fileTimePartName := strings.Split(fileName[1], fileEnd)
			if ok, err := s.checkFileNameTime(fileTimePartName[0]); err != nil {
				return nil, err
			} else if !ok {
				continue
			}
			response = append(response, path.Join(dir, fileInfo.Name()))
		}
	}
	return response, nil
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

func (s *searchLog) checkFileNameTime(timeString string) (bool, error) {
	timeInfo, err := time.Parse(fileLayout, timeString)
	if err != nil {
		return false, err
	}
	to := s.filter.to.AddDate(0, 0, 1)
	if (timeInfo.Before(to) ||
		timeInfo.Equal(to)) &&
		(timeInfo.After(s.filter.from) || timeInfo.Equal(s.filter.from)) {
		return true, nil
	}
	return false, nil
}
