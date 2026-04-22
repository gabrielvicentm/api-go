package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := loadAllowedOrigins()

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if isAllowedOrigin(origin, allowedOrigins) {
				c.Header("Access-Control-Allow-Origin", origin)
			}

			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Origin, Accept")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func loadAllowedOrigins() map[string]struct{} {
	rawOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	if strings.TrimSpace(rawOrigins) == "" {
		rawOrigins = strings.Join([]string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:8080",
			"http://127.0.0.1:8080",
			"http://localhost:8081",
			"http://127.0.0.1:8081",
		}, ",")
	}

	allowedOrigins := make(map[string]struct{})
	for _, origin := range strings.Split(rawOrigins, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}

		allowedOrigins[trimmed] = struct{}{}
	}

	return allowedOrigins
}

func isAllowedOrigin(origin string, allowedOrigins map[string]struct{}) bool {
	if _, ok := allowedOrigins["*"]; ok {
		return true
	}

	_, ok := allowedOrigins[origin]
	return ok
}
