// Package web contains a small web framework extension.
package web

import (
	"context"
	"errors"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
	"net/http"
	"os"
	"syscall"
	"time"
)

// A Handler is a type that handles a http request within our own little mini
// framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal

	// every handler regardless of what it does, is gonna wrapped with the middlewares specified in this field
	mw []Middleware
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	// since App represents an API, we use it as a pointer
	return &App{
		// ContextMux represents an API therefore use pointer semantics(*ContextMux)
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity
// issue is identified. This method issues a SIGTERM.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
/* Even though there's a method named Handle with the same signature that exists in the contextMux, we wanna overwrite that method, we implement
our own version of it.
Note: Here, we're essentially wrapping code around the handler.*/
func (a *App) Handle(method string, path string, handler Handler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	// h is the outer layer function(think of the onion)
	h := func(w http.ResponseWriter, r *http.Request) {
		v := Values{
			TraceID:    uuid.NewString(),
			Now:        time.Time{},
			StatusCode: 0,
		}

		ctx := context.WithValue(r.Context(), key, &v)

		/* The error returned here means:
		- the error handler(middleware) has returned it
		- or some other call in between the error handler and us has failed
		Both seem fairly serious. If error handler has returned sth it's pretty much a shutdown error. So validate it's a shutdown error
		and if it is, signal a shutdown.

		If an error ends up all the way back to the framework(web package), we're gonna run validateShutdown() func and if it was a shutdown error,
		we issue a sigterm to shut down the app.*/
		if err := handler(ctx, w, r); err != nil {
			if validateError(err) {
				a.SignalShutdown()

				return
			}
		}
	}

	a.ContextMux.Handle(method, path, h)
}

// validateError validates the error for special conditions that do not
// warrant an actual shutdown by the system.
func validateError(err error) bool {

	// Ignore syscall.EPIPE and syscall.ECONNRESET errors which occurs
	// when a write operation happens on the http.ResponseWriter that
	// has simultaneously been disconnected by the client (TCP
	// connections is broken). For instance, when large amounts of
	// data is being written or streamed to the client.
	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// https://gosamples.dev/broken-pipe/
	// https://gosamples.dev/connection-reset-by-peer/

	switch {
	case errors.Is(err, syscall.EPIPE):

		// Usually, you get the broken pipe error when you write to the connection after the
		// RST (TCP RST Flag) is sent.
		// The broken pipe is a TCP/IP error occurring when you write to a stream where the
		// other end (the peer) has closed the underlying connection. The first write to the
		// closed connection causes the peer to reply with an RST packet indicating that the
		// connection should be terminated immediately. The second write to the socket that
		// has already received the RST causes the broken pipe error.
		return false

	case errors.Is(err, syscall.ECONNRESET):

		// Usually, you get connection reset by peer error when you read from the
		// connection after the RST (TCP RST Flag) is sent.
		// The connection reset by peer is a TCP/IP error that occurs when the other end (peer)
		// has unexpectedly closed the connection. It happens when you send a packet from your
		// end, but the other end crashes and forcibly closes the connection with the RST
		// packet instead of the TCP FIN, which is used to close a connection under normal
		// circumstances.
		return false
	}

	return true
}
