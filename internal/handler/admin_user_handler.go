package handler

import (
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gin-gonic/gin"
)

type AdminUserHandler struct{}

func NewAdminUserHandler() *AdminUserHandler {
	return &AdminUserHandler{}
}

func (h *AdminUserHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/usuarios", h.List)
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
