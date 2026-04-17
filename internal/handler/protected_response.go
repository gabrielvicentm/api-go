package handler

import (
	"net/http"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gin-gonic/gin"
)

func respondProtected(c *gin.Context, scope, description string) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rota protegida autorizada com sucesso",
		"data": gin.H{
			"scope":       scope,
			"description": description,
			"authenticated_user": gin.H{
				"id":         claims.UserID,
				"email":      claims.Email,
				"role":       claims.Role,
				"actor_type": claims.ActorType,
			},
		},
	})
}
