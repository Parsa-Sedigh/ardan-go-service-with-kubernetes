package mid

import (
	"context"
	"fmt"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/metrics"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/web"
	"net/http"
	"runtime/debug"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
/* We returned a web.Middleware although we don't do anything in that middleware. This is just for consistency with other middlewares.

We're using a named return arg here. We're doing it so we can leverage closures.*/
func Panics() web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				/* If the value returned from recover() doesn't equal nil, it means a panic has happened. */
				if rec := recover(); rec != nil {
					trace := debug.Stack()

					/* Do not log the stacktrace HERE. The error handler(middleware) logs. Instead, here we construct an error with the
					stacktrace in it and then that error will be received by the error middleware.*/
					err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace))

					metrics.AddPanics(ctx)
				}
			}()

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
