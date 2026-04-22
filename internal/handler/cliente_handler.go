package handler

import (
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gin-gonic/gin"
)

type ClienteHandler struct {
	repo *repository.ClienteRepository
}

func NewClienteHandler(repo *repository.ClienteRepository) *ClienteHandler {
	return &ClienteHandler{repo: repo}
}

func (h *ClienteHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/clientes", h.List)
	group.POST("/clientes", h.Create)
	group.GET("/clientes/:id", h.Show)
	group.PUT("/clientes/:id", h.Update)
	group.DELETE("/clientes/:id", h.Delete)
}

func (h *ClienteHandler) List(c *gin.Context) {
	page, limit := parsePagination(c)

	items, total, err := h.repo.List(c.Request.Context(), domain.ClienteListFilter{
		Search: strings.TrimSpace(c.Query("search")),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar clientes")
		return
	}

	respondList(c, "Clientes listados com sucesso", items, page, limit, total)
}

func (h *ClienteHandler) Create(c *gin.Context) {
	var input domain.ClienteCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de cadastro invalidos", err)
		return
	}

	item, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao cadastrar cliente")
		return
	}

	respondSuccess(c, http.StatusCreated, "Cliente cadastrado com sucesso", item)
}

func (h *ClienteHandler) Show(c *gin.Context) {
	item, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar cliente")
		return
	}

	respondSuccess(c, http.StatusOK, "Cliente carregado com sucesso", item)
}

func (h *ClienteHandler) Update(c *gin.Context) {
	var input domain.ClienteUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de edicao invalidos", err)
		return
	}

	item, err := h.repo.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao atualizar cliente")
		return
	}

	respondSuccess(c, http.StatusOK, "Cliente atualizado com sucesso", item)
}

func (h *ClienteHandler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		respondDomainError(c, err, "Erro interno ao remover cliente")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cliente removido com sucesso"})
}
