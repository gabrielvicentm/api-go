package handler

import (
	"net/http"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gin-gonic/gin"
)

type AdminUserHandler struct {
	authService domain.AuthService
}

func NewAdminUserHandler(authService domain.AuthService) *AdminUserHandler {
	return &AdminUserHandler{authService: authService}
}

func (h *AdminUserHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/usuarios", h.List)
	group.POST("/usuarios/password-reset-token", h.GeneratePasswordResetToken)
}

func (h *AdminUserHandler) RegisterSuperadminRoutes(group *gin.RouterGroup) {
	group.GET("/usuarios", h.ListSuperadmin)
}

func (h *AdminUserHandler) List(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if ok && claims.Role == "superadmin" {
		h.ListSuperadmin(c)
		return
	}

	respondProtected(c, "admin.usuarios.list", "Listagem protegida de usuarios administrativos")
}

func (h *AdminUserHandler) ListSuperadmin(c *gin.Context) {
	respondProtected(c, "superadmin.usuarios.list", "Listagem protegida exclusiva para superadmin")
}

func (h *AdminUserHandler) GeneratePasswordResetToken(c *gin.Context) {
	var input domain.AdminGeneratePasswordResetTokenRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados invalidos para gerar token de reset", err)
		return
	}

	response, err := h.authService.GenerateAdminPasswordResetToken(c.Request.Context(), input)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			c.JSON(http.StatusNotFound, gin.H{"message": "Usuario administrativo nao encontrado"})
		case domain.ErrInactiveUser:
			c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Erro interno ao gerar token de reset"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token de reset gerado com sucesso",
		"data":    response,
	})
}
