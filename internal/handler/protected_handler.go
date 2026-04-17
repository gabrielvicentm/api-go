package handler

import (
	"net/http"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gin-gonic/gin"
)

type ProtectedHandler struct{}

func NewProtectedHandler() *ProtectedHandler {
	return &ProtectedHandler{}
}

func (h *ProtectedHandler) RegisterAdminRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin")
	admin.Use(
		authMiddleware,
		middleware.RequireActorTypes(domain.ActorTypeAdmin),
	)

	admin.GET("/dashboard", h.AdminDashboard)
	admin.GET("/motoristas", h.AdminMotoristasList)
	admin.POST("/motoristas", h.AdminMotoristasCreate)
	admin.GET("/veiculos", h.AdminVeiculosList)
	admin.POST("/veiculos", h.AdminVeiculosCreate)
	admin.GET("/viagens", h.AdminViagensList)
	admin.POST("/viagens", h.AdminViagensCreate)
	admin.GET("/usuarios", h.AdminUsuariosList)

	superadmin := admin.Group("/superadmin")
	superadmin.Use(middleware.RequireRoles("superadmin"))
	superadmin.GET("/usuarios", h.SuperadminUsuariosList)
}

func (h *ProtectedHandler) RegisterMotoristaRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	motorista := router.Group("/motorista")
	motorista.Use(
		authMiddleware,
		middleware.RequireActorTypes(domain.ActorTypeMotorista),
	)

	motorista.GET("/viagens", h.MotoristaViagensList)
	motorista.GET("/viagens/atual", h.MotoristaViagemAtual)
	motorista.POST("/ocorrencias", h.MotoristaOcorrenciasCreate)
	motorista.POST("/abastecimentos", h.MotoristaAbastecimentosCreate)
	motorista.POST("/viagens/finalizacao", h.MotoristaFinalizacaoCreate)
	motorista.GET("/notificacoes", h.MotoristaNotificacoesList)
}

func (h *ProtectedHandler) AdminDashboard(c *gin.Context) {
	h.respondProtected(c, "admin.dashboard.read", "Resumo do dashboard administrativo")
}

func (h *ProtectedHandler) AdminMotoristasList(c *gin.Context) {
	h.respondProtected(c, "admin.motoristas.list", "Listagem protegida de motoristas")
}

func (h *ProtectedHandler) AdminMotoristasCreate(c *gin.Context) {
	h.respondProtected(c, "admin.motoristas.create", "Cadastro protegido de motoristas")
}

func (h *ProtectedHandler) AdminVeiculosList(c *gin.Context) {
	h.respondProtected(c, "admin.veiculos.list", "Listagem protegida de veiculos")
}

func (h *ProtectedHandler) AdminVeiculosCreate(c *gin.Context) {
	h.respondProtected(c, "admin.veiculos.create", "Cadastro protegido de veiculos")
}

func (h *ProtectedHandler) AdminViagensList(c *gin.Context) {
	h.respondProtected(c, "admin.viagens.list", "Listagem protegida de viagens")
}

func (h *ProtectedHandler) AdminViagensCreate(c *gin.Context) {
	h.respondProtected(c, "admin.viagens.create", "Criacao protegida de viagens")
}

func (h *ProtectedHandler) AdminUsuariosList(c *gin.Context) {
	h.respondProtected(c, "admin.usuarios.list", "Listagem protegida de usuarios administrativos")
}

func (h *ProtectedHandler) SuperadminUsuariosList(c *gin.Context) {
	h.respondProtected(c, "superadmin.usuarios.list", "Listagem protegida exclusiva para superadmin")
}

func (h *ProtectedHandler) MotoristaViagensList(c *gin.Context) {
	h.respondProtected(c, "motorista.viagens.list", "Listagem protegida das viagens do motorista")
}

func (h *ProtectedHandler) MotoristaViagemAtual(c *gin.Context) {
	h.respondProtected(c, "motorista.viagens.atual", "Consulta protegida da viagem atual do motorista")
}

func (h *ProtectedHandler) MotoristaOcorrenciasCreate(c *gin.Context) {
	h.respondProtected(c, "motorista.ocorrencias.create", "Registro protegido de ocorrencias")
}

func (h *ProtectedHandler) MotoristaAbastecimentosCreate(c *gin.Context) {
	h.respondProtected(c, "motorista.abastecimentos.create", "Registro protegido de abastecimentos")
}

func (h *ProtectedHandler) MotoristaFinalizacaoCreate(c *gin.Context) {
	h.respondProtected(c, "motorista.viagens.finalizacao.create", "Solicitacao protegida de finalizacao de viagem")
}

func (h *ProtectedHandler) MotoristaNotificacoesList(c *gin.Context) {
	h.respondProtected(c, "motorista.notificacoes.list", "Listagem protegida de notificacoes do motorista")
}

func (h *ProtectedHandler) respondProtected(c *gin.Context, scope, description string) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rota protegida autorizada com sucesso",
		"data": gin.H{
			"scope":       scope,
			"description": description,
			"authenticated_user": gin.H{
				"id":         claims.UserID,
				"email":      claims.Email,
				"role":       claims.Role,
				"actor_type": claims.ActorType,
			},
		},
	})
}
