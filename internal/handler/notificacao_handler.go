package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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
	group.GET("/notificacoes/stream", h.StreamAdmin)
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

func (h *NotificacaoHandler) StreamAdmin(c *gin.Context) {
	h.streamByRecipient(c, domain.DestinatarioTipoAdmin)
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

func (h *NotificacaoHandler) streamByRecipient(c *gin.Context, destinatarioTipo string) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Streaming nao suportado"})
		return
	}

	subscription, err := h.service.Subscribe(destinatarioTipo, claims.UserID)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao conectar stream de notificacoes")
		return
	}
	defer subscription.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	c.SSEvent("connected", gin.H{"status": "ok"})
	flusher.Flush()

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case item, ok := <-subscription.Events:
			if !ok {
				return
			}
			c.SSEvent("notificacao", item)
			flusher.Flush()
		case <-heartbeat.C:
			c.SSEvent("ping", gin.H{"at": time.Now().UTC()})
			flusher.Flush()
		}
	}
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
