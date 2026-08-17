package http

import (
	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"github.com/yudadh/finance-bot-tracker-api/internal/http/handler"
)

func NewRouter(cfg config.AppConfig) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	healthHandler := handler.NewHealthHandler(cfg)

	api := router.Group("/api")

	{
		api.GET("/health", healthHandler.Show)
	}

	return router
}