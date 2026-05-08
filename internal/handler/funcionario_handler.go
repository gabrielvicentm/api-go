package handler

import (
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gin-gonic/gin"
)

type FuncionarioHandler struct {
	repo *repository.FuncionarioRepository
}

func NewFuncionarioHandler(repo *repository.FuncionarioRepository) *FuncionarioHandler {
	return &FuncionarioHandler{repo: repo}
}

func (h *FuncionarioHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/funcionarios", h.List)
	group.POST("/funcionarios", h.Create)
	group.GET("/funcionarios/:id", h.Show)
	group.PUT("/funcionarios/:id", h.Update)
	group.DELETE("/funcionarios/:id", h.Delete)
	group.PATCH("/funcionarios/:id/status", h.UpdateStatus)
}

func (h *FuncionarioHandler) List(c *gin.Context) {
	page, limit := parsePagination(c)

	includeMotorista := true
	if raw := strings.TrimSpace(strings.ToLower(c.Query("include_motoristas"))); raw == "false" || raw == "0" {
		includeMotorista = false
	}

	items, total, err := h.repo.List(c.Request.Context(), domain.FuncionarioListFilter{
		Search:           strings.TrimSpace(c.Query("search")),
		Status:           strings.TrimSpace(c.Query("status")),
		Tipo:             strings.TrimSpace(c.Query("tipo")),
		Page:             page,
		Limit:            limit,
		IncludeMotorista: includeMotorista,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar funcionarios")
		return
	}

	respondList(c, "Funcionarios listados com sucesso", items, page, limit, total)
}

func (h *FuncionarioHandler) Create(c *gin.Context) {
	var input domain.FuncionarioCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de cadastro invalidos", err)
		return
	}

	item, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao cadastrar funcionario")
		return
	}

	respondSuccess(c, http.StatusCreated, "Funcionario cadastrado com sucesso", item)
}

func (h *FuncionarioHandler) Show(c *gin.Context) {
	item, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar funcionario")
		return
	}

	respondSuccess(c, http.StatusOK, "Funcionario carregado com sucesso", item)
}

func (h *FuncionarioHandler) Update(c *gin.Context) {
	var input domain.FuncionarioUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de edicao invalidos", err)
		return
	}

	item, err := h.repo.Update(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao atualizar funcionario")
		return
	}

	respondSuccess(c, http.StatusOK, "Funcionario atualizado com sucesso", item)
}

func (h *FuncionarioHandler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		respondDomainError(c, err, "Erro interno ao remover funcionario")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Funcionario removido com sucesso"})
}

func (h *FuncionarioHandler) UpdateStatus(c *gin.Context) {
	var input domain.FuncionarioStatusUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Status invalido", err)
		return
	}

	item, err := h.repo.UpdateStatus(c.Request.Context(), c.Param("id"), input.Status)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao atualizar status do funcionario")
		return
	}

	respondSuccess(c, http.StatusOK, "Status do funcionario atualizado com sucesso", item)
}
