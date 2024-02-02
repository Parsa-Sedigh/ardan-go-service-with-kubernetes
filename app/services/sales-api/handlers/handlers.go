package handlers

import (
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/app/services/sales-api/handlers/v1/testgrp"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/app/services/sales-api/handlers/v1/usergrp"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/core/user"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/core/user/stores/userdb"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/web/auth"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/web/v1/mid"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"os"
)

// APIMuxConfig contains all the mandatory systems required by handlers
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	DB       *sqlx.DB
}

// APIMux constructs a http.Handler with all application routes defined
func APIMux(cfg APIMuxConfig) *web.App {
	// Panics() should always be the last middleware, so it's as close to the handler as possible
	app := web.NewApp(cfg.Shutdown, mid.Logger(cfg.Log), mid.Errors(cfg.Log), mid.Metrics(), mid.Panics())

	// bind a route to the mux(app variable). If a req with GET method comes in, execute this handler
	app.Handle(http.MethodGet, "/test", testgrp.Test)
	app.Handle(http.MethodGet, "/test/auth", testgrp.Test, mid.Authenticate(cfg.Auth), mid.Authorize(cfg.Auth, auth.RuleAdminOnly))

	// =============================================================================

	usrCore := user.NewCore(userdb.NewStore(cfg.Log, cfg.DB))

	ugh := usergrp.New(usrCore, cfg.Auth)

	app.Handle(http.MethodGet, "/users", ugh.Query)

	return app
}
