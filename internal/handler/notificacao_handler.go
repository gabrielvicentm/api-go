package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
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
	h.listByRecipient(c, domain.DestinatarioTipoAdmin)
}

func (h *NotificacaoHandler) MarkAsReadAdmin(c *gin.Context) {
	h.markAsRead(c, domain.DestinatarioTipoAdmin)
}

func (h *NotificacaoHandler) ListMotorista(c *gin.Context) {
	h.listByRecipient(c, domain.DestinatarioTipoMotorista)
}

func (h *NotificacaoHandler) MarkAsReadMotorista(c *gin.Context) {
	h.markAsRead(c, domain.DestinatarioTipoMotorista)
}

func (h *NotificacaoHandler) listByRecipient(c *gin.Context, destinatarioTipo string) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	page, limit := parsePagination(c)
	lida, err := parseOptionalBoolQuery(c, "lida")
	if err != nil {
		respondError(c, http.StatusBadRequest, "Parametro lida invalido", err)
		return
	}

	items, total, err := h.service.ListByRecipient(c.Request.Context(), domain.NotificacaoListFilter{
		DestinatarioTipo: destinatarioTipo,
		DestinatarioID:   claims.UserID,
		Lida:             lida,
		Page:             page,
		Limit:            limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar notificacoes")
		return
	}

	respondList(c, "Notificacoes listadas com sucesso", items, page, limit, total)
}

func (h *NotificacaoHandler) markAsRead(c *gin.Context, destinatarioTipo string) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	item, err := h.service.MarkAsRead(c.Request.Context(), c.Param("id"), destinatarioTipo, claims.UserID)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao marcar notificacao como lida")
		return
	}

	respondSuccess(c, http.StatusOK, "Notificacao marcada como lida com sucesso", item)
}

func parseOptionalBoolQuery(c *gin.Context, key string) (*bool, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
