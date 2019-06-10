package logging

import (
	journal "github.com/integration-system/isp-journal"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/logger"
)

func WithLogging(journal journal.Journal, enable bool) backend.PostProcessor {
	return func(ctx backend.RequestCtx) {
		if !enable {
			return
		}

		err := ctx.Error()
		if err != nil {
			if err := journal.Error(ctx.Method(), ctx.RequestBody(), ctx.ResponseBody(), err); err != nil {
				logger.Warnf("could not write to file journal: %v", err)
			}
		} else {
			if err := journal.Info(ctx.Method(), ctx.RequestBody(), ctx.ResponseBody()); err != nil {
				logger.Warnf("could not write to file journal: %v", err)
			}
		}
	}
}
