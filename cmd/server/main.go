package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/app"
	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"github.com/yudadh/finance-bot-tracker-api/internal/logger"
)

func main() {
	cfg := config.NewConfig()

	if err := cfg.ReadInConfig(); err != nil {
		log.Fatal("failed to read config", err)
		os.Exit(1)
	}
	
	appConfig := config.New(cfg)
	log := logger.New(appConfig.App.AppEnv)
	application, err := app.New(appConfig, log)
	if err != nil {
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	
	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			log.Error(err.Error())
			return
		}
		
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		log.Error(err.Error())
	}
}