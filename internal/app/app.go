package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"github.com/yudadh/finance-bot-tracker-api/internal/database"
	appHttp "github.com/yudadh/finance-bot-tracker-api/internal/http"
	"gorm.io/gorm"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger
	server *http.Server
	db     *gorm.DB
}

func New(cfg config.Config, logger *slog.Logger) *App {
	database := database.NewGormDB(cfg.Database)
	router := appHttp.NewRouter(cfg.App)

	server := &http.Server{
		Addr:              cfg.App.HTTPAddress(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		server: server,
		db: database,
	}
}

// Run starts the HTTP server and blocks until the server stops.
// It returns an error if the server fails to start or stops unexpectedly,
// except when the server is closed normally.
func (a *App) Run() error {
	a.logger.Info("starting http server", "addr", a.server.Addr, "env", a.cfg.App.AppEnv)

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Shutdown stops the HTTP server, it returns an error if the server fails to stop
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down http server")

	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	a.logger.Info("http server stopped")
	return nil
}
