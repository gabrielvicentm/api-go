package handler

import (
	"errors"
	"net/http"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service        domain.AuthService
	authMiddleware gin.HandlerFunc
}

func NewAuthHandler(service domain.AuthService, authMiddleware gin.HandlerFunc) *AuthHandler {
	return &AuthHandler{
		service:        service,
		authMiddleware: authMiddleware,
	}
}

func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	auth := router.Group("/auth")
	auth.POST("/login", h.LoginAdmin)
	auth.POST("/admin/login", h.LoginAdmin)
	auth.POST("/motorista/login", h.LoginMotorista)
	auth.POST("/refresh", h.RefreshToken)
	auth.POST("/logout", h.Logout)

	protected := auth.Group("")
	protected.Use(h.authMiddleware)
	protected.GET("/me", h.Me)
	protected.POST("/change-password", h.ChangePassword)
}

func (h *AuthHandler) LoginAdmin(c *gin.Context) {
	var input domain.AdminLoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de login invalidos", err)
		return
	}

	response, err := h.service.LoginAdmin(c.Request.Context(), input)
	if err != nil {
		h.handleAuthError(c, err, "Erro interno ao realizar login do admin")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login do admin realizado com sucesso",
		"data":    response,
	})
}

func (h *AuthHandler) LoginMotorista(c *gin.Context) {
	var input domain.MotoristaLoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de login invalidos", err)
		return
	}

	response, err := h.service.LoginMotorista(c.Request.Context(), input)
	if err != nil {
		h.handleAuthError(c, err, "Erro interno ao realizar login do motorista")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login do motorista realizado com sucesso",
		"data":    response,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input domain.RefreshTokenRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Refresh token invalido", err)
		return
	}

	response, err := h.service.RefreshToken(c.Request.Context(), input)
	if err != nil {
		h.handleAuthError(c, err, "Erro interno ao renovar token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token renovado com sucesso",
		"data":    response,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var input domain.LogoutRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Refresh token invalido", err)
		return
	}

	if err := h.service.Logout(c.Request.Context(), input); err != nil {
		h.handleAuthError(c, err, "Erro interno ao realizar logout")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout realizado com sucesso"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	user, err := h.service.GetProfile(c.Request.Context(), claims.ActorType, claims.UserID)
	if err != nil {
		h.handleAuthError(c, err, "Erro interno ao buscar usuario autenticado")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usuario autenticado carregado com sucesso",
		"data":    user,
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	var input domain.ChangePasswordRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de troca de senha invalidos", err)
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), claims.ActorType, claims.UserID, input); err != nil {
		h.handleAuthError(c, err, "Erro interno ao trocar senha")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Senha alterada com sucesso",
	})
}

func (h *AuthHandler) handleAuthError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
	case errors.Is(err, domain.ErrInactiveUser):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	case errors.Is(err, domain.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
	case errors.Is(err, domain.ErrExpiredToken):
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": fallbackMessage})
	}
}

func respondError(c *gin.Context, status int, message string, err error) {
	c.JSON(status, gin.H{
		"message": message,
		"error":   err.Error(),
	})
}
