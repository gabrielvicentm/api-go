package handler

import "github.com/gin-gonic/gin"

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/dashboard", h.ShowDashboard)
	group.GET("/alertas", h.ListAlerts)
}

func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	respondProtected(c, "admin.dashboard.read", "Resumo protegido do dashboard administrativo")
}

func (h *DashboardHandler) ListAlerts(c *gin.Context) {
	respondProtected(c, "admin.alertas.list", "Listagem protegida de alertas operacionais")
}
