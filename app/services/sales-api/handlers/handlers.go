package handlers

import (
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/app/services/sales-api/handlers/v1/testgrp"
	"github.com/dimfeld/httptreemux/v5"
	"go.uber.org/zap"
	"net/http"
	"os"
)

// APIMuxConfig contains all the mandatory systems required by handlers
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
}

// APIMux constructs a http.Handler with all application routes defined
func APIMux(cfg APIMuxConfig) http.Handler {
	mux := httptreemux.NewContextMux()

	// bind a route to the mux. If a req with GET method comes in, execute this handler
	mux.Handle(http.MethodGet, "/test", testgrp.Test)

	return mux
}
