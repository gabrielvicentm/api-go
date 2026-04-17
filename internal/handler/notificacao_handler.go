package handler

import "github.com/gin-gonic/gin"

type NotificacaoHandler struct{}

func NewNotificacaoHandler() *NotificacaoHandler {
	return &NotificacaoHandler{}
}

func (h *NotificacaoHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/notificacoes", h.ListAdmin)
	group.PATCH("/notificacoes/:id/lida", h.MarkAsReadAdmin)
}

func (h *NotificacaoHandler) RegisterMotoristaRoutes(group *gin.RouterGroup) {
	group.GET("/notificacoes", h.ListMotorista)
	group.PATCH("/notificacoes/:id/lida", h.MarkAsReadMotorista)
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
