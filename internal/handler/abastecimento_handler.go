package handler

import "github.com/gin-gonic/gin"

type AbastecimentoHandler struct{}

func NewAbastecimentoHandler() *AbastecimentoHandler {
	return &AbastecimentoHandler{}
}

func (h *AbastecimentoHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/abastecimentos", h.ListAdmin)
	group.GET("/abastecimentos/:id", h.ShowAdmin)
}

func (h *AbastecimentoHandler) RegisterMotoristaRoutes(group *gin.RouterGroup) {
	group.POST("/abastecimentos", h.Create)
	group.GET("/abastecimentos", h.ListMotorista)
}

func (h *AbastecimentoHandler) ListAdmin(c *gin.Context) {
	respondProtected(c, "admin.abastecimentos.list", "Listagem protegida de abastecimentos")
}

func (h *AbastecimentoHandler) ShowAdmin(c *gin.Context) {
	respondProtected(c, "admin.abastecimentos.read", "Consulta protegida de detalhes do abastecimento")
}

func (h *AbastecimentoHandler) Create(c *gin.Context) {
	respondProtected(c, "motorista.abastecimentos.create", "Registro protegido de abastecimentos")
}

func (h *AbastecimentoHandler) ListMotorista(c *gin.Context) {
	respondProtected(c, "motorista.abastecimentos.list", "Listagem protegida de abastecimentos do motorista")
}
