package handler

import "github.com/gin-gonic/gin"

type ClienteHandler struct{}

func NewClienteHandler() *ClienteHandler {
	return &ClienteHandler{}
}

func (h *ClienteHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/clientes", h.List)
	group.POST("/clientes", h.Create)
	group.GET("/clientes/:id", h.Show)
	group.PUT("/clientes/:id", h.Update)
	group.DELETE("/clientes/:id", h.Delete)
}

func (h *ClienteHandler) List(c *gin.Context) {
	respondProtected(c, "admin.clientes.list", "Listagem protegida de clientes")
}

func (h *ClienteHandler) Create(c *gin.Context) {
	respondProtected(c, "admin.clientes.create", "Cadastro protegido de clientes")
}

func (h *ClienteHandler) Show(c *gin.Context) {
	respondProtected(c, "admin.clientes.read", "Consulta protegida de cliente")
}

func (h *ClienteHandler) Update(c *gin.Context) {
	respondProtected(c, "admin.clientes.update", "Edicao protegida de cliente")
}

func (h *ClienteHandler) Delete(c *gin.Context) {
	respondProtected(c, "admin.clientes.delete", "Remocao protegida de cliente")
}
