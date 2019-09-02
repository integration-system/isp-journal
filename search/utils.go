package search

import (
	"github.com/integration-system/isp-lib/logger"
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

func findAllMatchedFiles(filter Filter, baseDir string) ([]string, error) {
	if dirs, err := findDirs(filter, baseDir); err != nil {
		return nil, err
	} else if len(dirs) == 0 {
		return nil, nil
	} else {
		return findFiles(dirs, filter, baseDir)
	}
}

func findDirs(filter Filter, baseDir string) ([]string, error) {
	f := time.Date(
		filter.from.Year(), filter.from.Month(), filter.from.Day(),
		0, 0, 0, 0, filter.from.Location())
	t := time.Date(
		filter.to.Year(), filter.to.Month(), filter.to.Day(),
		0, 0, 0, 0, filter.to.Location()).AddDate(0, 0, 1)
	dirs := make([]string, 0)
	if arrayDateDir, err := ioutil.ReadDir(baseDir); err != nil {
		return nil, err
	} else {
		for _, dateDirInfo := range arrayDateDir {
			if dirName, err := time.Parse(dirLayout, dateDirInfo.Name()); err != nil {
				continue
			} else if (dirName.After(f) || dirName.Equal(f)) && (dirName.Before(t) || dirName.Equal(t)) {
				dirs = append(dirs, dateDirInfo.Name())
			}
		}
	}
	return dirs, nil
}

func findFiles(dirs []string, filter Filter, baseDir string) ([]string, error) {
	response := make([]string, 0)
	for _, dir := range dirs {
		dir := path.Join(baseDir, dir, filter.moduleName)
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
			if !filter.checkHost(fileName[0]) {
				continue
			}
			if len(fileName) < 2 {
				logger.Warnf("invalid file name '%s'", fileInfo.Name())
				continue
			}
			fileTimePartName := strings.Split(fileName[1], fileEnd)
			if ok, err := checkFileNameTime(fileTimePartName[0], filter); err != nil {
				return nil, err
			} else if !ok {
				continue
			}
			response = append(response, path.Join(dir, fileInfo.Name()))
		}
	}
	return response, nil
}

func checkFileNameTime(timeString string, filter Filter) (bool, error) {
	timeInfo, err := time.Parse(fileLayout, timeString)
	if err != nil {
		return false, err
	}
	to := filter.to.AddDate(0, 0, 1)
	if (timeInfo.Before(to) ||
		timeInfo.Equal(to)) &&
		(timeInfo.After(filter.from) || timeInfo.Equal(filter.from)) {
		return true, nil
	}
	return false, nil
}
