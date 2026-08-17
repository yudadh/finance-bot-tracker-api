package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/config"
)

type HealthHandler struct {
	cfg config.AppConfig
}

func NewHealthHandler(cfg config.AppConfig) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

func (h *HealthHandler) Show(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"service":     h.cfg.AppName,
		"environment": h.cfg.AppEnv,
		"time":        time.Now().UTC().Format(time.RFC3339),
	})
}