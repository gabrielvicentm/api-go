package middleware

import (
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/security"
	"github.com/gin-gonic/gin"
)

const ContextAccessClaimsKey = "access_claims"

func AuthMiddleware(tokenManager *security.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "cabecalho Authorization obrigatorio"})
			return
		}

		tokenParts := strings.SplitN(authorizationHeader, " ", 2)
		if len(tokenParts) != 2 || !strings.EqualFold(tokenParts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "formato do token invalido"})
			return
		}

		claims, err := tokenManager.ValidateAccessToken(tokenParts[1])
		if err != nil {
			status := http.StatusUnauthorized
			if err != domain.ErrInvalidToken && err != domain.ErrExpiredToken {
				status = http.StatusInternalServerError
			}

			c.AbortWithStatusJSON(status, gin.H{"message": err.Error()})
			return
		}

		c.Set(ContextAccessClaimsKey, claims)
		c.Next()
	}
}

func GetAccessClaims(c *gin.Context) (*domain.AccessTokenClaims, bool) {
	claimsValue, exists := c.Get(ContextAccessClaimsKey)
	if !exists {
		return nil, false
	}

	claims, ok := claimsValue.(*domain.AccessTokenClaims)
	if !ok {
		return nil, false
	}

	return claims, true
}

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		claimsValue, exists := c.Get(ContextAccessClaimsKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
			return
		}

		claims, ok := claimsValue.(*domain.AccessTokenClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
			return
		}

		if _, ok := allowed[claims.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": domain.ErrForbidden.Error()})
			return
		}

		c.Next()
	}
}

func RequireActorTypes(allowedTypes ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedTypes))
	for _, actorType := range allowedTypes {
		allowed[actorType] = struct{}{}
	}

	return func(c *gin.Context) {
		claimsValue, exists := c.Get(ContextAccessClaimsKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
			return
		}

		claims, ok := claimsValue.(*domain.AccessTokenClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
			return
		}

		if _, ok := allowed[claims.ActorType]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": domain.ErrForbidden.Error()})
			return
		}

		c.Next()
	}
}
