package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"github.com/yudadh/finance-bot-tracker-api/internal/security"
)

func AdminAuth(cfg config.JWTConfig) gin.HandlerFunc {
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

		c.Set("admin_user_id", claims.AdminUserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}