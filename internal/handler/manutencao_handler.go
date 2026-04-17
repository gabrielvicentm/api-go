package handler

import "github.com/gin-gonic/gin"

type ManutencaoHandler struct{}

func NewManutencaoHandler() *ManutencaoHandler {
	return &ManutencaoHandler{}
}

func (h *ManutencaoHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/manutencoes", h.List)
	group.POST("/manutencoes", h.Create)
	group.GET("/manutencoes/:id", h.Show)
	group.PUT("/manutencoes/:id", h.Update)
	group.GET("/veiculos/:id/manutencoes", h.ListByVehicle)
}

func (h *ManutencaoHandler) List(c *gin.Context) {
	respondProtected(c, "admin.manutencoes.list", "Listagem protegida de manutencoes")
}

func (h *ManutencaoHandler) Create(c *gin.Context) {
	respondProtected(c, "admin.manutencoes.create", "Cadastro protegido de manutencao")
}

func (h *ManutencaoHandler) Show(c *gin.Context) {
	respondProtected(c, "admin.manutencoes.read", "Consulta protegida de manutencao")
}

func (h *ManutencaoHandler) Update(c *gin.Context) {
	respondProtected(c, "admin.manutencoes.update", "Edicao protegida de manutencao")
}

func (h *ManutencaoHandler) ListByVehicle(c *gin.Context) {
	respondProtected(c, "admin.veiculos.manutencoes.list", "Historico protegido de manutencoes por veiculo")
}
