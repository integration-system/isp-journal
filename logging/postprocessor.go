package logging

import (
	journal "github.com/integration-system/isp-journal"
	"github.com/integration-system/isp-journal/codes"
	"github.com/integration-system/isp-lib/v2/backend"
	log "github.com/integration-system/isp-log"
)

func WithLogging(journal journal.Journal, enable bool, includeMethods ...string) backend.PostProcessor {
	includedMethods := make(map[string]bool, len(includeMethods))
	for _, m := range includeMethods {
		includedMethods[m] = true
	}
	return func(ctx backend.RequestCtx) {
		if !enable {
			return
		}

		method := ctx.Method()
		if include := includedMethods[method]; include {
			err := ctx.Error()
			if err != nil {
				if err := journal.Error(method, ctx.RequestBody(), ctx.ResponseBody(), err); err != nil {
					log.Warnf(codes.JournalingError, "could not write to file journal: %v", err)
				}
			} else {
				if err := journal.Info(method, ctx.RequestBody(), ctx.ResponseBody()); err != nil {
					log.Warnf(codes.JournalingError, "could not write to file journal: %v", err)
				}
			}
		}
	}
}
