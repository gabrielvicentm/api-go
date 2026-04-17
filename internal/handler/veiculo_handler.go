package handler

import "github.com/gin-gonic/gin"

type VeiculoHandler struct{}

func NewVeiculoHandler() *VeiculoHandler {
	return &VeiculoHandler{}
}

func (h *VeiculoHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/veiculos", h.List)
	group.POST("/veiculos", h.Create)
	group.GET("/veiculos/:id", h.Show)
	group.PUT("/veiculos/:id", h.Update)
	group.GET("/veiculos/:id/custos", h.Costs)
	group.GET("/veiculos/:id/consumo", h.Consumption)
	group.GET("/veiculos/:id/historico", h.History)
}

func (h *VeiculoHandler) List(c *gin.Context) {
	respondProtected(c, "admin.veiculos.list", "Listagem protegida de veiculos")
}

func (h *VeiculoHandler) Create(c *gin.Context) {
	respondProtected(c, "admin.veiculos.create", "Cadastro protegido de veiculos")
}

func (h *VeiculoHandler) Show(c *gin.Context) {
	respondProtected(c, "admin.veiculos.read", "Consulta protegida de detalhes do veiculo")
}

func (h *VeiculoHandler) Update(c *gin.Context) {
	respondProtected(c, "admin.veiculos.update", "Edicao protegida de veiculo")
}

func (h *VeiculoHandler) Costs(c *gin.Context) {
	respondProtected(c, "admin.veiculos.costs.read", "Leitura protegida de custo total por veiculo")
}

func (h *VeiculoHandler) Consumption(c *gin.Context) {
	respondProtected(c, "admin.veiculos.consumption.read", "Leitura protegida de consumo medio por veiculo")
}

func (h *VeiculoHandler) History(c *gin.Context) {
	respondProtected(c, "admin.veiculos.history.read", "Leitura protegida do historico completo do veiculo")
}
