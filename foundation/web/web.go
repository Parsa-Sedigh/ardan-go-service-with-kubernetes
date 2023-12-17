// Package web contains a small web framework extension.
package web

import (
	"context"
	"fmt"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
	"net/http"
	"os"
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

		if err := handler(ctx, w, r); err != nil {
			fmt.Print(err)

			return
		}
	}

	a.ContextMux.Handle(method, path, h)
}
