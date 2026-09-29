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
	categoryService handler.CategoryService,
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
	categoryHandler := handler.NewCategoryHandler(categoryService)

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
		// category
		category := api.Group("/category")
		category.Use(middleware.AdminAuth(cfg.JWT, userAdminService, logger))
		{
			category.POST("", handler.Handle(logger, categoryHandler.Create))
			category.GET("/:id", handler.Handle(logger, categoryHandler.FindByID))
			category.GET("", handler.Handle(logger, categoryHandler.FindAll))
			category.PUT("/:id", handler.Handle(logger, categoryHandler.Update))
			category.DELETE("/:id", handler.Handle(logger, categoryHandler.HardDelete))
		}

		api.GET("/health", healthHandler.Show)
	}

	return router
}