package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/security"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type adminUserFinder interface {
	FindByID(ctx context.Context, id uint64) (*service.FindByIDResult, error)
}

func AdminAuth(
	cfg config.JWTConfig,
	adminUserService adminUserFinder,
	logger *slog.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context)  {
		header := c.GetHeader("Authorization")

		jwtToken, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(jwtToken) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": true,
				"message": "unauthorized",
			})
			return
		}

		claims, err := security.ParseAdminToken(cfg, jwtToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": true,
				"message": "unauthorized",
			})
			return
		}

		admin, err := adminUserService.FindByID(c.Request.Context(), claims.AdminUserID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": true,
					"message": "unauthorized",
				})
				return
			}

			logger.ErrorContext(
				c.Request.Context(),
				"admin user lookup failed on middleware",
				"admin_user_id",
				claims.AdminUserID,
			)

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"message": "internal server error",
			})
			return
		}

		if admin.Status != domain.AdminUserStatusActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": true,
				"message": "inactive account",
			})
			return
		}

		c.Set("admin_user_id", claims.AdminUserID)
		c.Set("email", admin.Email)

		c.Next()
	}
}