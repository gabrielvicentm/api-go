package handler

import "github.com/gin-gonic/gin"

type OcorrenciaHandler struct{}

func NewOcorrenciaHandler() *OcorrenciaHandler {
	return &OcorrenciaHandler{}
}

func (h *OcorrenciaHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/ocorrencias", h.ListAdmin)
	group.GET("/ocorrencias/:id", h.ShowAdmin)
}

func (h *OcorrenciaHandler) RegisterMotoristaRoutes(group *gin.RouterGroup) {
	group.POST("/ocorrencias", h.Create)
	group.GET("/ocorrencias", h.ListMotorista)
}

func (h *OcorrenciaHandler) ListAdmin(c *gin.Context) {
	respondProtected(c, "admin.ocorrencias.list", "Listagem protegida de ocorrencias")
}

func (h *OcorrenciaHandler) ShowAdmin(c *gin.Context) {
	respondProtected(c, "admin.ocorrencias.read", "Consulta protegida de detalhes da ocorrencia")
}

func (h *OcorrenciaHandler) Create(c *gin.Context) {
	respondProtected(c, "motorista.ocorrencias.create", "Registro protegido de ocorrencias")
}

func (h *OcorrenciaHandler) ListMotorista(c *gin.Context) {
	respondProtected(c, "motorista.ocorrencias.list", "Listagem protegida de ocorrencias do motorista")
}
