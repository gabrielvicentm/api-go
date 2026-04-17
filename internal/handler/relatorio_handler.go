package handler

import "github.com/gin-gonic/gin"

type RelatorioHandler struct{}

func NewRelatorioHandler() *RelatorioHandler {
	return &RelatorioHandler{}
}

func (h *RelatorioHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/relatorios/viagens", h.Trips)
	group.GET("/relatorios/combustivel", h.Fuel)
	group.GET("/relatorios/manutencoes", h.Maintenance)
	group.GET("/relatorios/custos", h.Costs)
	group.GET("/relatorios/desempenho", h.Performance)
	group.GET("/relatorios/lucro-por-viagem", h.ProfitPerTrip)
	group.GET("/relatorios/exportacoes/xlsx", h.ExportXLSX)
	group.GET("/relatorios/exportacoes/csv", h.ExportCSV)
}

func (h *RelatorioHandler) Trips(c *gin.Context) {
	respondProtected(c, "admin.relatorios.viagens.read", "Agregacao protegida de relatorio de viagens")
}

func (h *RelatorioHandler) Fuel(c *gin.Context) {
	respondProtected(c, "admin.relatorios.combustivel.read", "Agregacao protegida de relatorio de combustivel")
}

func (h *RelatorioHandler) Maintenance(c *gin.Context) {
	respondProtected(c, "admin.relatorios.manutencoes.read", "Agregacao protegida de relatorio de manutencoes")
}

func (h *RelatorioHandler) Costs(c *gin.Context) {
	respondProtected(c, "admin.relatorios.custos.read", "Agregacao protegida de relatorio de custos")
}

func (h *RelatorioHandler) Performance(c *gin.Context) {
	respondProtected(c, "admin.relatorios.desempenho.read", "Agregacao protegida de relatorio de desempenho")
}

func (h *RelatorioHandler) ProfitPerTrip(c *gin.Context) {
	respondProtected(c, "admin.relatorios.lucro_por_viagem.read", "Agregacao protegida de relatorio de lucro por viagem")
}

func (h *RelatorioHandler) ExportXLSX(c *gin.Context) {
	respondProtected(c, "admin.relatorios.export.xlsx", "Exportacao protegida de relatorios em XLSX")
}

func (h *RelatorioHandler) ExportCSV(c *gin.Context) {
	respondProtected(c, "admin.relatorios.export.csv", "Exportacao protegida de relatorios em CSV")
}
