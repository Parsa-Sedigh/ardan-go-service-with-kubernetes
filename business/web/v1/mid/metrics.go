package mid

import (
	"context"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/metrics"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/web"
	"net/http"
)

// Metrics updates program counters.
func Metrics() web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// put the metrics data in to the ctx
			ctx = metrics.Set(ctx)

			err := handler(ctx, w, r)

			n := metrics.AddRequests(ctx)

			/* polling. Because we could have 1000s of goroutines coming through. So there's one request every so often(n%1000 == 0)
			that gets hit with a little bit more work of capturing and updating the number of goroutines using AddGoroutines() .*/
			if n%1000 == 0 {
				metrics.AddGoroutines(ctx)
			}

			// we're still behind the error handling(error handler middleware is a parent of this middleware), but we can update the metrics for errors
			if err != nil {
				metrics.AddErrors(ctx)
			}

			return err
		}

		return h
	}

	return m
}
