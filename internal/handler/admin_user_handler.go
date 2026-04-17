package handler

import "github.com/gin-gonic/gin"

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
	respondProtected(c, "admin.usuarios.list", "Listagem protegida de usuarios administrativos")
}

func (h *AdminUserHandler) ListSuperadmin(c *gin.Context) {
	respondProtected(c, "superadmin.usuarios.list", "Listagem protegida exclusiva para superadmin")
}
