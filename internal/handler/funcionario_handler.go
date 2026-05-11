package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gabrielvicentm/api-go.git/internal/service"
	"github.com/gin-gonic/gin"
)

type FuncionarioHandler struct {
	repo         *repository.FuncionarioRepository
	photoStorage service.PhotoStorage
}

func NewFuncionarioHandler(repo *repository.FuncionarioRepository, photoStorage service.PhotoStorage) *FuncionarioHandler {
	return &FuncionarioHandler{
		repo:         repo,
		photoStorage: photoStorage,
	}
}

func (h *FuncionarioHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/funcionarios", h.List)
	group.POST("/funcionarios", h.Create)
	group.GET("/funcionarios/:id", h.Show)
	group.PUT("/funcionarios/:id", h.Update)
	group.DELETE("/funcionarios/:id", h.Delete)
	group.PATCH("/funcionarios/:id/status", h.UpdateStatus)
	group.POST("/funcionarios/:id/foto", h.UploadPhoto)
	group.GET("/folha-pagamento", h.ListPayroll)
	group.GET("/funcionarios/:id/folha-pagamento", h.ShowPayroll)
	group.PUT("/funcionarios/:id/folha-pagamento", h.UpsertPayroll)
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

func (h *FuncionarioHandler) UploadPhoto(c *gin.Context) {
	if h.photoStorage == nil {
		respondDomainError(c, fmt.Errorf("photo storage nao configurado"), "Armazenamento de fotos nao configurado")
		return
	}

	file, err := c.FormFile("foto")
	if err != nil {
		respondError(c, http.StatusBadRequest, "Arquivo de foto obrigatorio", err)
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		respondDomainError(c, err, "Erro interno ao abrir foto do funcionario")
		return
	}
	defer openedFile.Close()

	contentType, err := detectImageContentType(openedFile)
	if err != nil {
		respondError(c, http.StatusBadRequest, "Arquivo enviado nao e uma imagem valida", err)
		return
	}

	photoURL, err := h.photoStorage.UploadFuncionarioPhoto(c.Request.Context(), openedFile, file.Filename, contentType)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao enviar foto do funcionario")
		return
	}

	item, err := h.repo.UpdatePhoto(c.Request.Context(), c.Param("id"), photoURL)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao vincular foto ao funcionario")
		return
	}

	respondSuccess(c, http.StatusOK, "Foto do funcionario enviada com sucesso", item)
}

func (h *FuncionarioHandler) ListPayroll(c *gin.Context) {
	items, err := h.repo.ListFolhaPagamento(c.Request.Context(), domain.FolhaPagamentoListFilter{
		Search:      strings.TrimSpace(c.Query("search")),
		Status:      strings.TrimSpace(c.Query("status")),
		Competencia: strings.TrimSpace(c.Query("competencia")),
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar a folha de pagamento")
		return
	}

	respondSuccess(c, http.StatusOK, "Folha de pagamento carregada com sucesso", items)
}

func (h *FuncionarioHandler) ShowPayroll(c *gin.Context) {
	item, err := h.repo.GetFolhaPagamento(c.Request.Context(), c.Param("id"), strings.TrimSpace(c.Query("competencia")))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao carregar a folha do funcionario")
		return
	}

	respondSuccess(c, http.StatusOK, "Folha do funcionario carregada com sucesso", item)
}

func (h *FuncionarioHandler) UpsertPayroll(c *gin.Context) {
	var input domain.FolhaPagamentoUpsertRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados da folha invalidos", err)
		return
	}

	item, err := h.repo.UpsertFolhaPagamento(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao salvar a folha do funcionario")
		return
	}

	respondSuccess(c, http.StatusOK, "Folha do funcionario salva com sucesso", item)
}
