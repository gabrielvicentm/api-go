package handler

import (
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	repo *repository.DashboardRepository
}

func NewDashboardHandler(repo *repository.DashboardRepository) *DashboardHandler {
	return &DashboardHandler{repo: repo}
}

func (h *DashboardHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/dashboard", h.ShowDashboard)
	group.GET("/alertas", h.ListAlerts)
}

func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	snapshot, err := h.repo.GetSnapshot(c.Request.Context())
	if err != nil {
		respondDomainError(c, err, "Erro interno ao carregar dashboard")
		return
	}

	respondSuccess(c, 200, "Dashboard carregado com sucesso", snapshot)
}

func (h *DashboardHandler) ListAlerts(c *gin.Context) {
	respondProtected(c, "admin.alertas.list", "Listagem protegida de alertas operacionais")
}
