package search

import (
	"github.com/integration-system/isp-journal/entry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type SearchRequest struct {
	ModuleName string `valid:"required~Required"`
	From       time.Time
	To         time.Time
	Host       []string
	Event      []string
	Level      []string
	Limit      int `valid:"required~Required,range(1|10000)"`
	Offset     int
}

type Filter struct {
	hostByExist  map[string]bool
	eventByExist map[string]bool
	levelByExist map[string]bool

	from time.Time
	to   time.Time
}

func NewFilter(req SearchRequest) (Filter, error) {
	f := Filter{}
	for _, value := range req.Host {
		f.hostByExist[value] = true
	}
	for _, value := range req.Event {
		f.eventByExist[value] = true
	}
	for _, value := range req.Level {
		f.levelByExist[value] = true
	}

	if err := f.defineTimeForSearch(req.From, req.To); err != nil {
		return f, err
	}

	return f, nil
}

func (f *Filter) defineTimeForSearch(from, to time.Time) error {
	if from.IsZero() {
		from = time.Now().UTC().AddDate(0, 0, -1)
	} else {
		from = from.UTC()
	}

	if to.IsZero() {
		to = time.Now().UTC()
	} else {
		to = to.UTC()
		if to.Before(from) {
			return status.Error(codes.InvalidArgument, "expected FROM will before TO")
		}
	}
	f.from = from
	f.to = to
	return nil
}

func (f *Filter) checkTimeField(timeString string) (bool, error) {
	timeInfo, err := entry.ParserTime(timeString)
	if err != nil {
		return false, err
	}
	if (timeInfo.Before(f.to) || timeInfo.Equal(f.to)) && (timeInfo.After(f.from) || timeInfo.Equal(f.from)) {
		return true, nil
	}
	return false, nil
}

func (f *Filter) checkEntry(entries *entry.Entry) bool {
	if !f.checkLevel(entries.Level) {
		return false
	}
	if !f.checkEvent(entries.Event) {
		return false
	}
	if !f.checkHost(entries.Host) {
		return false
	}
	return true
}

func (f *Filter) checkLevel(level string) bool {
	return f.checkEntryField(f.levelByExist, level)
}

func (f *Filter) checkEvent(event string) bool {
	return f.checkEntryField(f.eventByExist, event)
}

func (f *Filter) checkHost(host string) bool {
	return f.checkEntryField(f.hostByExist, host)
}

func (f *Filter) checkEntryField(expected map[string]bool, field string) bool {
	if len(expected) == 0 {
		return true
	} else if expected[field] {
		return true
	}
	return false
}
