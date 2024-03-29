package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/app/services/sales-api/handlers"
	database "github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/sys/database/pgx"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/web/auth"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/web/v1/debug"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/keystore"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/logger"
	"github.com/ardanlabs/conf/v3"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var build = "develop"

func main() {
	// construct the logger
	log, err := logger.New("SALES-API")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0), "BUILD-", build)

	// -------------------------------------------------------------------------
	// Configuration

	/* Why we're not using a named type here instead of a literal struct?
	It's actually important to leverage a literal struct here. Because we don't want to create a master config type to be passed around your program.
	We just want to pass the individual fields of this struct around. But not naming this into a type, I enforce this idea that this
	shouldn't be passed around as a whole instead we should pass the individual fields.*/
	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
		}
		DB struct {
			User     string `conf:"default:postgres"`
			Password string `conf:"default:postgres,mask"`

			/* if you have telepresence, use this default. One big win with telepresence is all of our defaults work whether we're inside
			or outside the cluster.*/
			//Host         string `conf:"default:database-service.sales-system.svc.cluster.local"`
			// if you don't have telepresence, use this one
			Host         string `conf:"default:localhost"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:2"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		Auth struct {
			KeysFolder string `conf:"default:zarf/keys/"`
			ActiveKID  string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
			Issuer     string `conf:"default:service project"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	const prefix = "SALES"
	/* Since the conf package also handles the command line flags, if it returns an error, you can check to see if the error was ErrHelpWanted and then
	you can print the help information that came out.*/
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}

	// show the config that you're running
	log.Infow("startup", "config", out)

	// -------------------------------------------------------------------------
	// Database Support

	log.Infow("startup", "status", "initializing database support", "host", cfg.DB.Host)

	db, err := database.Open(database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()

	// -------------------------------------------------------------------------
	// Initialize authentication support

	log.Infow("startup", "status", "initializing authentication support")

	// Simple keystore versus using Vault.
	ks, err := keystore.NewFS(os.DirFS(cfg.Auth.KeysFolder))
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	//vault, err := vault.New(vault.Config{
	//	Address:   cfg.Vault.Address,
	//	Token:     cfg.Vault.Token,
	//	MountPath: cfg.Vault.MountPath,
	//})
	//if err != nil {
	//	return fmt.Errorf("constructing vault: %w", err)
	//}

	authCfg := auth.Config{
		Log:       log,
		KeyLookup: ks,
	}

	auth, err := auth.New(authCfg)
	if err != nil {
		return fmt.Errorf("constructing auth: %w", err)
	}

	// -------------------------------------------------------------------------
	// Start Debug Service

	log.Infow("startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

	/* create a goroutine that blocks on a ListenAndServe() call on whatever IP and port is for debug and the second parameter is a mux that
	registers all of the routes for these handlers.*/
	// this is an orp
	go func() {
		/* do not use http.DefaultServeMux here. Some dependency or yourself, could expose some endpoints that shoudldn't be used
		by anyone and should be behind a firewall, but you accidentally exposed them by binding the http.DefaultServeMux directly.
		Instead, create your own mux and bind that here instead of http.DefaultServeMux.*/
		/* debug.Mux() registers the /debug routes which include of pprof routes and readiness and liveness handlers. */
		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux(build, log, db)); err != nil {
			log.Errorw("shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Infow("startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
		Auth:     auth,
		DB:       db,
	})

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux, // mux
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown completed", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// -------------------------------------------------------------------------
		// LOAD SHEDDING FOR APPLICATION API
		/* tell the api to start load shedding and we give this operation a timeout using ctx. We give the api some time to make sure
		that all the goroutines terminate, but this can go on forever, so we give it a timeout.*/
		if err := api.Shutdown(ctx); err != nil {
			api.Close()

			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
