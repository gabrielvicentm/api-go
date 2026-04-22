package handler

import (
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gin-gonic/gin"
)

type VeiculoHandler struct {
	repo *repository.VeiculoRepository
}

func NewVeiculoHandler(repo *repository.VeiculoRepository) *VeiculoHandler {
	return &VeiculoHandler{repo: repo}
}

func (h *VeiculoHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/veiculos", h.List)
	group.POST("/veiculos", h.Create)
	group.GET("/veiculos/:id", h.Show)
	group.PUT("/veiculos/:id", h.Update)
	group.DELETE("/veiculos/:id", h.Delete)
	group.GET("/veiculos/:id/custos", h.Costs)
	group.GET("/veiculos/:id/consumo", h.Consumption)
	group.GET("/veiculos/:id/historico", h.History)
}

func (h *VeiculoHandler) List(c *gin.Context) {
	page, limit := parsePagination(c)

	items, total, err := h.repo.List(c.Request.Context(), domain.VeiculoListFilter{
		Search: strings.TrimSpace(c.Query("search")),
		Status: strings.TrimSpace(c.Query("status")),
		Tipo:   strings.TrimSpace(c.Query("tipo")),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar veiculos")
		return
	}

	respondList(c, "Veiculos listados com sucesso", items, page, limit, total)
}

func (h *VeiculoHandler) Create(c *gin.Context) {
	var input domain.VeiculoCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de cadastro invalidos", err)
		return
	}

	item, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao cadastrar veiculo")
		return
	}

	respondSuccess(c, http.StatusCreated, "Veiculo cadastrado com sucesso", item)
}

func (h *VeiculoHandler) Show(c *gin.Context) {
	item, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar veiculo")
		return
	}

	respondSuccess(c, http.StatusOK, "Veiculo carregado com sucesso", item)
}

func (h *VeiculoHandler) Update(c *gin.Context) {
	var input domain.VeiculoUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de edicao invalidos", err)
		return
	}

	item, err := h.repo.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao atualizar veiculo")
		return
	}

	respondSuccess(c, http.StatusOK, "Veiculo atualizado com sucesso", item)
}

func (h *VeiculoHandler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		respondDomainError(c, err, "Erro interno ao remover veiculo")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Veiculo removido com sucesso"})
}

func (h *VeiculoHandler) Costs(c *gin.Context) {
	item, err := h.repo.GetCosts(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar custos do veiculo")
		return
	}

	respondSuccess(c, http.StatusOK, "Custos do veiculo carregados com sucesso", item)
}

func (h *VeiculoHandler) Consumption(c *gin.Context) {
	item, err := h.repo.GetConsumption(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar consumo do veiculo")
		return
	}

	respondSuccess(c, http.StatusOK, "Consumo do veiculo carregado com sucesso", item)
}

func (h *VeiculoHandler) History(c *gin.Context) {
	items, err := h.repo.GetHistory(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar historico do veiculo")
		return
	}

	respondSuccess(c, http.StatusOK, "Historico do veiculo carregado com sucesso", items)
}
