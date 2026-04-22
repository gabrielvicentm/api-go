package handler

import (
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gin-gonic/gin"
)

type TipoCargaHandler struct {
	repo *repository.TipoCargaRepository
}

func NewTipoCargaHandler(repo *repository.TipoCargaRepository) *TipoCargaHandler {
	return &TipoCargaHandler{repo: repo}
}

func (h *TipoCargaHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/tipos-carga", h.List)
	group.POST("/tipos-carga", h.Create)
	group.GET("/tipos-carga/:id", h.Show)
	group.PUT("/tipos-carga/:id", h.Update)
	group.DELETE("/tipos-carga/:id", h.Delete)
}

func (h *TipoCargaHandler) List(c *gin.Context) {
	page, limit := parsePagination(c)

	items, total, err := h.repo.List(c.Request.Context(), domain.TipoCargaListFilter{
		Search: strings.TrimSpace(c.Query("search")),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar tipos de carga")
		return
	}

	respondList(c, "Tipos de carga listados com sucesso", items, page, limit, total)
}

func (h *TipoCargaHandler) Create(c *gin.Context) {
	var input domain.TipoCargaCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de cadastro invalidos", err)
		return
	}

	item, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao cadastrar tipo de carga")
		return
	}

	respondSuccess(c, http.StatusCreated, "Tipo de carga cadastrado com sucesso", item)
}

func (h *TipoCargaHandler) Show(c *gin.Context) {
	item, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar tipo de carga")
		return
	}

	respondSuccess(c, http.StatusOK, "Tipo de carga carregado com sucesso", item)
}

func (h *TipoCargaHandler) Update(c *gin.Context) {
	var input domain.TipoCargaUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de edicao invalidos", err)
		return
	}

	item, err := h.repo.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao atualizar tipo de carga")
		return
	}

	respondSuccess(c, http.StatusOK, "Tipo de carga atualizado com sucesso", item)
}

func (h *TipoCargaHandler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		respondDomainError(c, err, "Erro interno ao remover tipo de carga")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipo de carga removido com sucesso"})
}
