package handler

import (
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gin-gonic/gin"
)

type ManutencaoHandler struct {
	repo *repository.ManutencaoRepository
}

func NewManutencaoHandler(repo *repository.ManutencaoRepository) *ManutencaoHandler {
	return &ManutencaoHandler{repo: repo}
}

func (h *ManutencaoHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/manutencoes", h.List)
	group.POST("/manutencoes", h.Create)
	group.GET("/manutencoes/:id", h.Show)
	group.PUT("/manutencoes/:id", h.Update)
	group.GET("/veiculos/:id/manutencoes", h.ListByVehicle)
}

func (h *ManutencaoHandler) List(c *gin.Context) {
	page, limit := parsePagination(c)

	items, total, err := h.repo.List(c.Request.Context(), domain.ManutencaoListFilter{
		Search:    strings.TrimSpace(c.Query("search")),
		Status:    strings.TrimSpace(c.Query("status")),
		Tipo:      strings.TrimSpace(c.Query("tipo")),
		VeiculoID: strings.TrimSpace(c.Query("veiculo_id")),
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar manutencoes")
		return
	}

	respondList(c, "Manutencoes listadas com sucesso", items, page, limit, total)
}

func (h *ManutencaoHandler) Create(c *gin.Context) {
	var input domain.ManutencaoCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de cadastro invalidos", err)
		return
	}

	item, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao cadastrar manutencao")
		return
	}

	respondSuccess(c, http.StatusCreated, "Manutencao cadastrada com sucesso", item)
}

func (h *ManutencaoHandler) Show(c *gin.Context) {
	item, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar manutencao")
		return
	}

	respondSuccess(c, http.StatusOK, "Manutencao carregada com sucesso", item)
}

func (h *ManutencaoHandler) Update(c *gin.Context) {
	var input domain.ManutencaoUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de edicao invalidos", err)
		return
	}

	item, err := h.repo.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao atualizar manutencao")
		return
	}

	respondSuccess(c, http.StatusOK, "Manutencao atualizada com sucesso", item)
}

func (h *ManutencaoHandler) ListByVehicle(c *gin.Context) {
	page, limit := parsePagination(c)

	items, total, err := h.repo.List(c.Request.Context(), domain.ManutencaoListFilter{
		VeiculoID: c.Param("id"),
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar manutencoes do veiculo")
		return
	}

	respondList(c, "Manutencoes do veiculo listadas com sucesso", items, page, limit, total)
}
