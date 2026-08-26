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
	"github.com/yudadh/finance-bot-tracker-api/internal/parser"
	"github.com/yudadh/finance-bot-tracker-api/internal/repository"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
	"github.com/yudadh/finance-bot-tracker-api/internal/telegram"
	"gorm.io/gorm"
)

type App struct {
	cfg         config.Config
	logger      *slog.Logger
	server      *http.Server
	db          *gorm.DB
	telegramBot *telegram.Bot
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	database := database.NewGormDB(cfg.Database)
	router := appHttp.NewRouter(cfg.App)

	// repository
	transactionRepository := repository.NewTransactionRepository(database)
	categoryRepository := repository.NewCategoryRepository(database)
	parserAttempRepository := repository.NewParserAttemptRepository(database)
	userRepository := repository.NewUserRepository(database)
	budgetRepository := repository.NewBudgetRepository(database)

	// parser
	parser := parser.NewRuleBasedParser()

	transactionService := service.NewTransactionService(
		transactionRepository,
		categoryRepository,
		parserAttempRepository,
		parser,
		logger,
	)
	userService := service.NewUserService(userRepository, logger)
	budgetService := service.NewBudgetService(budgetRepository, transactionRepository, logger)

	botHandler := telegram.NewBotHandler(transactionService, userService, budgetService, logger)
	bot, err := telegram.NewBot(cfg.App.TelegramBotToken, botHandler, logger)
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Addr:              cfg.App.HTTPAddress(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		cfg:         cfg,
		logger:      logger,
		server:      server,
		db:          database,
		telegramBot: bot,
	}, nil
}

// Run starts the HTTP server with telegram bot and blocks until the server stops.
// It returns an error if the server fails to start or stops unexpectedly,
// except when the server is closed normally.
func (a *App) Run(ctx context.Context) error {
	a.logger.Info("starting telegram bot")
	go a.telegramBot.Start(ctx)

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
