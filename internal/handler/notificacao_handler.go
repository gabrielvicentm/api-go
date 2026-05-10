package handler

import (
	"net/http"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/service"
	"github.com/gin-gonic/gin"
)

type NotificacaoHandler struct {
	service *service.NotificacaoService
}

func NewNotificacaoHandler(service *service.NotificacaoService) *NotificacaoHandler {
	return &NotificacaoHandler{service: service}
}

func (h *NotificacaoHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/notificacoes", h.ListAdmin)
	group.PATCH("/notificacoes/:id/lida", h.MarkAsReadAdmin)
}

func (h *NotificacaoHandler) RegisterMotoristaRoutes(group *gin.RouterGroup) {
	group.GET("/notificacoes", h.ListMotorista)
	group.PATCH("/notificacoes/:id/lida", h.MarkAsReadMotorista)
}

func (h *NotificacaoHandler) RegisterInternalRoutes(group *gin.RouterGroup) {
	group.POST("/notificacoes", h.CreateInternal)
}

func (h *NotificacaoHandler) CreateInternal(c *gin.Context) {
	var input domain.NotificacaoCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de cadastro invalidos", err)
		return
	}

	item, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao cadastrar notificacao")
		return
	}

	respondSuccess(c, http.StatusCreated, "Notificacao cadastrada com sucesso", item)
}

func (h *NotificacaoHandler) ListAdmin(c *gin.Context) {
	respondProtected(c, "admin.notificacoes.list", "Listagem protegida de notificacoes administrativas")
}

func (h *NotificacaoHandler) MarkAsReadAdmin(c *gin.Context) {
	respondProtected(c, "admin.notificacoes.read.update", "Marcacao protegida de notificacao administrativa como lida")
}

func (h *NotificacaoHandler) ListMotorista(c *gin.Context) {
	respondProtected(c, "motorista.notificacoes.list", "Listagem protegida de notificacoes do motorista")
}

func (h *NotificacaoHandler) MarkAsReadMotorista(c *gin.Context) {
	respondProtected(c, "motorista.notificacoes.read.update", "Marcacao protegida de notificacao do motorista como lida")
}
