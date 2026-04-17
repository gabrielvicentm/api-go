package handler

import "github.com/gin-gonic/gin"

type MotoristaHandler struct{}

func NewMotoristaHandler() *MotoristaHandler {
	return &MotoristaHandler{}
}

func (h *MotoristaHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/motoristas", h.ListAdmin)
	group.POST("/motoristas", h.Create)
	group.GET("/motoristas/:id", h.ShowAdmin)
	group.PUT("/motoristas/:id", h.Update)
	group.PATCH("/motoristas/:id/status", h.UpdateStatus)
	group.GET("/motoristas/:id/indicadores", h.Indicators)
	group.GET("/motoristas/:id/viagens", h.TripsHistory)
	group.GET("/motoristas/:id/ocorrencias", h.OccurrencesHistory)
}

func (h *MotoristaHandler) RegisterMotoristaRoutes(group *gin.RouterGroup) {
	group.GET("/perfil", h.ShowSelf)
	group.GET("/viagens", h.ListOwnTrips)
	group.GET("/viagens/atual", h.CurrentTrip)
	group.GET("/viagens/historico", h.History)
}

func (h *MotoristaHandler) ListAdmin(c *gin.Context) {
	respondProtected(c, "admin.motoristas.list", "Listagem protegida de motoristas")
}

func (h *MotoristaHandler) Create(c *gin.Context) {
	respondProtected(c, "admin.motoristas.create", "Cadastro protegido de motoristas")
}

func (h *MotoristaHandler) ShowAdmin(c *gin.Context) {
	respondProtected(c, "admin.motoristas.read", "Consulta protegida de detalhes do motorista")
}

func (h *MotoristaHandler) Update(c *gin.Context) {
	respondProtected(c, "admin.motoristas.update", "Edicao protegida de motorista")
}

func (h *MotoristaHandler) UpdateStatus(c *gin.Context) {
	respondProtected(c, "admin.motoristas.status.update", "Atualizacao protegida de status do motorista")
}

func (h *MotoristaHandler) Indicators(c *gin.Context) {
	respondProtected(c, "admin.motoristas.indicators.read", "Leitura protegida de indicadores do motorista")
}

func (h *MotoristaHandler) TripsHistory(c *gin.Context) {
	respondProtected(c, "admin.motoristas.trips.list", "Historico protegido de viagens do motorista")
}

func (h *MotoristaHandler) OccurrencesHistory(c *gin.Context) {
	respondProtected(c, "admin.motoristas.occurrences.list", "Historico protegido de ocorrencias do motorista")
}

func (h *MotoristaHandler) ShowSelf(c *gin.Context) {
	respondProtected(c, "motorista.profile.read", "Consulta protegida do proprio perfil do motorista")
}

func (h *MotoristaHandler) ListOwnTrips(c *gin.Context) {
	respondProtected(c, "motorista.viagens.list", "Listagem protegida das viagens do motorista")
}

func (h *MotoristaHandler) CurrentTrip(c *gin.Context) {
	respondProtected(c, "motorista.viagens.atual", "Consulta protegida da viagem atual do motorista")
}

func (h *MotoristaHandler) History(c *gin.Context) {
	respondProtected(c, "motorista.viagens.historico", "Historico protegido de viagens anteriores do motorista")
}
