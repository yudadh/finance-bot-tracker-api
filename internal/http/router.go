package http

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"github.com/yudadh/finance-bot-tracker-api/internal/http/handler"
	"github.com/yudadh/finance-bot-tracker-api/internal/http/middleware"
)

func NewRouter(
	cfg config.AppConfig, 
	userAdminService handler.UserAdminService,
	userService handler.UserService,
	transactionService handler.TransactionService,
	logger *slog.Logger,
) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	healthHandler := handler.NewHealthHandler(cfg)
	userAdminHandler := handler.NewUserAdminHandler(userAdminService, userService, logger)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	api := router.Group("/api")

	{	
		api.POST("/admin/login", handler.Handle(logger, userAdminHandler.Login))
		admin := api.Group("/admin")
		admin.Use(middleware.AdminAuth(cfg.JWT, userAdminService, logger))
		{
			admin.GET("/me", handler.Handle(logger, userAdminHandler.GetMe))
			admin.GET("/users", handler.Handle(logger, userAdminHandler.GetAllUsers))
			admin.GET(
				"/transactions/users/:id", 
				handler.Handle(
					logger, 
					transactionHandler.GetTransactionsByUserPeriod,
				),
			)
		}
		api.GET("/health", healthHandler.Show)
	}

	return router
}