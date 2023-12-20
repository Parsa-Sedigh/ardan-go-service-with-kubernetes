package handlers

import (
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/app/services/sales-api/handlers/v1/testgrp"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/web/v1/mid"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/web"
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
func APIMux(cfg APIMuxConfig) *web.App {
	// Panics() should always be the last middleware, so it's as close to the handler as possible
	app := web.NewApp(cfg.Shutdown, mid.Logger(cfg.Log), mid.Errors(cfg.Log), mid.Metrics(), mid.Panics())

	// bind a route to the mux(app variable). If a req with GET method comes in, execute this handler
	app.Handle(http.MethodGet, "/test", testgrp.Test)

	return app
}
