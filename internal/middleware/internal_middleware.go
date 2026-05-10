package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func InternalAuthMiddlewareFromEnv() gin.HandlerFunc {
	expectedToken := strings.TrimSpace(os.Getenv("INTERNAL_API_TOKEN"))

	return func(c *gin.Context) {
		if expectedToken == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "token interno nao configurado"})
			return
		}

		token := strings.TrimSpace(c.GetHeader("X-Internal-Token"))
		if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "token interno invalido"})
			return
		}

		c.Next()
	}
}
