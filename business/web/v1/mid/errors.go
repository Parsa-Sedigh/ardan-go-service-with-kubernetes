package mid

import (
	"context"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/sys/validate"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/web/auth"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/web/v1"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/web"
	"go.uber.org/zap"
	"net/http"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
// Error handling means logging the error, so we needed to pass the logger.
// Note: Middlewares accepts the thing they need like the logger.
func Errors(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := handler(ctx, w, r); err != nil {
				log.Errorw("ERROR", "trace_id", web.GetTraceID(ctx), "message", err)

				var er v1.ErrorResponse
				var status int

				switch {
				case validate.IsFieldErrors(err):
					fieldErrors := validate.GetFieldErrors(err)
					er = v1.ErrorResponse{
						Error:  "data validation error",
						Fields: fieldErrors.Fields(),
					}
					status = http.StatusBadRequest

				case v1.IsRequestError(err):
					reqErr := v1.GetRequestError(err)
					er = v1.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				case auth.IsAuthError(err):
					er = v1.ErrorResponse{
						Error: http.StatusText(http.StatusUnauthorized),
					}
					status = http.StatusUnauthorized

				default:
					er = v1.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				/* If we receive the shutdown err we need to return it back to the base handler to shut down the service.
				So if we received a ShutdownError, we allow the shutdown to go back to the web framework. We can do some more
				analysis here as well.Here's a situation where the error handler is handling the error but it's still allowing
				the error to return back! But who is it returning the error back to? There's only one place that can still
				handle this and that's in foundation/web/web.go .*/
				if web.IsShutdown(err) {
					return err
				}
			}

			return nil
		}

		return h
	}

	return m
}
