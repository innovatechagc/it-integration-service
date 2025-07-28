package middleware

import (
	"net/http"

	"github.com/company/microservice-template/internal/auth"
	"github.com/gin-gonic/gin"
)

func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := jwtManager.ExtractTokenFromHeader(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": err.Error(),
				"data":    nil,
			})
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_TOKEN",
				"message": "Invalid or expired token",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// Agregar claims al contexto
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)
		c.Next()
	}
}

func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "User roles not found",
				"data":    nil,
			})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "Invalid user roles format",
				"data":    nil,
			})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range userRoles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "INSUFFICIENT_PERMISSIONS",
				"message": "Insufficient permissions for this resource",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func SwaggerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Solo permitir Swagger en desarrollo
		if gin.Mode() == gin.ReleaseMode {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "NOT_FOUND",
				"message": "Resource not found",
				"data":    nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}