package mid

import (
	"context"
	"fmt"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/web"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// with Logger, we have the ability to do the logging before and after the handler that was called
func Logger(log *zap.SugaredLogger) web.Middleware {
	// the idea is we want to execute the handler that was passed in but with some extra work

	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v := web.GetValues(ctx)

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Infow("request started", "trace_id", v.TraceID, "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr)

			err := handler(ctx, w, r)

			// by seeing this log, it means we don't have leaking(the goroutine created for the req is gonna complete) and also it's not blocked,
			// because we actually saw this log
			log.Infow("request completed", "trace_id", v.TraceID, "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr,
				"statuscode", v.StatusCode, "since", time.Since(v.Now).String())

			return err
		}

		return h
	}

	return m
}
