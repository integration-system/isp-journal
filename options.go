package journal

import (
	"github.com/integration-system/isp-journal/log"
)

type Option func(journal *fileJournal)

func WithAfterRotation(callback func(log log.LogFile)) Option {
	return func(journal *fileJournal) {
		journal.afterRotation = callback
	}
}
