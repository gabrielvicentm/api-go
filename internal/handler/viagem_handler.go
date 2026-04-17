package handler

import "github.com/gin-gonic/gin"

type ViagemHandler struct{}

func NewViagemHandler() *ViagemHandler {
	return &ViagemHandler{}
}

func (h *ViagemHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/viagens", h.ListAdmin)
	group.POST("/viagens", h.Create)
	group.GET("/viagens/:id", h.ShowAdmin)
	group.PUT("/viagens/:id", h.Update)
	group.GET("/viagens/:id/historico", h.History)
	group.GET("/viagens/:id/documentos", h.DocumentsList)
	group.POST("/viagens/:id/documentos", h.DocumentsUpload)
	group.GET("/viagens/:id/finalizacoes", h.FinalizationsListAdmin)
	group.POST("/viagens/:id/finalizacoes/:finalizacaoId/aprovar", h.ApproveFinalization)
	group.POST("/viagens/:id/finalizacoes/:finalizacaoId/rejeitar", h.RejectFinalization)
}

func (h *ViagemHandler) RegisterMotoristaRoutes(group *gin.RouterGroup) {
	group.GET("/viagens", h.ListMotorista)
	group.GET("/viagens/atual", h.Current)
	group.GET("/viagens/historico", h.HistoryMotorista)
	group.POST("/viagens/paradas", h.CreateStop)
	group.POST("/viagens/finalizacao", h.RequestFinalization)
}

func (h *ViagemHandler) ListAdmin(c *gin.Context) {
	respondProtected(c, "admin.viagens.list", "Listagem protegida de viagens com filtros")
}

func (h *ViagemHandler) Create(c *gin.Context) {
	respondProtected(c, "admin.viagens.create", "Criacao protegida de viagens")
}

func (h *ViagemHandler) ShowAdmin(c *gin.Context) {
	respondProtected(c, "admin.viagens.read", "Consulta protegida de detalhes da viagem")
}

func (h *ViagemHandler) Update(c *gin.Context) {
	respondProtected(c, "admin.viagens.update", "Edicao protegida de viagem")
}

func (h *ViagemHandler) History(c *gin.Context) {
	respondProtected(c, "admin.viagens.history.read", "Leitura protegida do historico de alteracoes da viagem")
}

func (h *ViagemHandler) DocumentsList(c *gin.Context) {
	respondProtected(c, "admin.viagens.documents.list", "Listagem protegida dos documentos da viagem")
}

func (h *ViagemHandler) DocumentsUpload(c *gin.Context) {
	respondProtected(c, "admin.viagens.documents.create", "Upload protegido de documentos da viagem")
}

func (h *ViagemHandler) FinalizationsListAdmin(c *gin.Context) {
	respondProtected(c, "admin.viagens.finalizations.list", "Listagem protegida de solicitacoes de finalizacao")
}

func (h *ViagemHandler) ApproveFinalization(c *gin.Context) {
	respondProtected(c, "admin.viagens.finalizations.approve", "Aprovacao protegida de finalizacao de viagem")
}

func (h *ViagemHandler) RejectFinalization(c *gin.Context) {
	respondProtected(c, "admin.viagens.finalizations.reject", "Rejeicao protegida de finalizacao de viagem")
}

func (h *ViagemHandler) ListMotorista(c *gin.Context) {
	respondProtected(c, "motorista.viagens.list", "Listagem protegida das viagens do motorista")
}

func (h *ViagemHandler) Current(c *gin.Context) {
	respondProtected(c, "motorista.viagens.atual", "Consulta protegida da viagem atual do motorista")
}

func (h *ViagemHandler) HistoryMotorista(c *gin.Context) {
	respondProtected(c, "motorista.viagens.historico", "Historico protegido de viagens anteriores do motorista")
}

func (h *ViagemHandler) CreateStop(c *gin.Context) {
	respondProtected(c, "motorista.viagens.paradas.create", "Registro protegido de parada da viagem")
}

func (h *ViagemHandler) RequestFinalization(c *gin.Context) {
	respondProtected(c, "motorista.viagens.finalizacao.create", "Solicitacao protegida de finalizacao de viagem")
}
