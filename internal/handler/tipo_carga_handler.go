package handler

import "github.com/gin-gonic/gin"

type TipoCargaHandler struct{}

func NewTipoCargaHandler() *TipoCargaHandler {
	return &TipoCargaHandler{}
}

func (h *TipoCargaHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/tipos-carga", h.List)
	group.POST("/tipos-carga", h.Create)
	group.PUT("/tipos-carga/:id", h.Update)
	group.DELETE("/tipos-carga/:id", h.Delete)
}

func (h *TipoCargaHandler) List(c *gin.Context) {
	respondProtected(c, "admin.tipos_carga.list", "Listagem protegida de tipos de carga")
}

func (h *TipoCargaHandler) Create(c *gin.Context) {
	respondProtected(c, "admin.tipos_carga.create", "Cadastro protegido de tipo de carga")
}

func (h *TipoCargaHandler) Update(c *gin.Context) {
	respondProtected(c, "admin.tipos_carga.update", "Edicao protegida de tipo de carga")
}

func (h *TipoCargaHandler) Delete(c *gin.Context) {
	respondProtected(c, "admin.tipos_carga.delete", "Remocao protegida de tipo de carga")
}
