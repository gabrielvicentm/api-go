package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gabrielvicentm/api-go.git/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type MotoristaHandler struct {
	repo         *repository.MotoristaRepository
	photoStorage service.PhotoStorage
}

func NewMotoristaHandler(repo *repository.MotoristaRepository, photoStorage service.PhotoStorage) *MotoristaHandler {
	return &MotoristaHandler{
		repo:         repo,
		photoStorage: photoStorage,
	}
}

func (h *MotoristaHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/motoristas", h.ListAdmin)
	group.POST("/motoristas", h.Create)
	group.GET("/motoristas/:id", h.ShowAdmin)
	group.PUT("/motoristas/:id", h.Update)
	group.DELETE("/motoristas/:id", h.Delete)
	group.PATCH("/motoristas/:id/status", h.UpdateStatus)
	group.POST("/motoristas/:id/foto", h.UploadPhoto)
	group.GET("/motoristas/:id/indicadores", h.Indicators)
	group.GET("/motoristas/:id/viagens", h.TripsHistory)
	group.GET("/motoristas/:id/ocorrencias", h.OccurrencesHistory)
}

func (h *MotoristaHandler) RegisterMotoristaRoutes(group *gin.RouterGroup) {
	group.GET("/perfil", h.ShowSelf)
}

func (h *MotoristaHandler) ListAdmin(c *gin.Context) {
	page, limit := parsePagination(c)

	items, total, err := h.repo.List(c.Request.Context(), domain.MotoristaListFilter{
		Search: strings.TrimSpace(c.Query("search")),
		Status: strings.TrimSpace(c.Query("status")),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar motoristas")
		return
	}

	respondList(c, "Motoristas listados com sucesso", items, page, limit, total)
}

func (h *MotoristaHandler) Create(c *gin.Context) {
	var input domain.MotoristaCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de cadastro invalidos", err)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.SenhaInicial), bcrypt.DefaultCost)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao gerar senha do motorista")
		return
	}

	item, err := h.repo.Create(c.Request.Context(), input, string(passwordHash))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao cadastrar motorista")
		return
	}

	respondSuccess(c, http.StatusCreated, "Motorista cadastrado com sucesso", item)
}

func (h *MotoristaHandler) ShowAdmin(c *gin.Context) {
	item, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar motorista")
		return
	}

	respondSuccess(c, http.StatusOK, "Motorista carregado com sucesso", item)
}

func (h *MotoristaHandler) Update(c *gin.Context) {
	var input domain.MotoristaUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de edicao invalidos", err)
		return
	}

	var passwordHash *string
	if strings.TrimSpace(input.NovaSenha) != "" {
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(input.NovaSenha), bcrypt.DefaultCost)
		if err != nil {
			respondDomainError(c, err, "Erro interno ao atualizar senha do motorista")
			return
		}
		hash := string(hashBytes)
		passwordHash = &hash
	}

	item, err := h.repo.Update(c.Request.Context(), c.Param("id"), input, passwordHash)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao atualizar motorista")
		return
	}

	respondSuccess(c, http.StatusOK, "Motorista atualizado com sucesso", item)
}

func (h *MotoristaHandler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		respondDomainError(c, err, "Erro interno ao remover motorista")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Motorista removido com sucesso"})
}

func (h *MotoristaHandler) UpdateStatus(c *gin.Context) {
	var input domain.MotoristaStatusUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Status invalido", err)
		return
	}

	item, err := h.repo.UpdateStatus(c.Request.Context(), c.Param("id"), input.Status)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao atualizar status do motorista")
		return
	}

	respondSuccess(c, http.StatusOK, "Status do motorista atualizado com sucesso", item)
}

func (h *MotoristaHandler) UploadPhoto(c *gin.Context) {
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
		respondDomainError(c, err, "Erro interno ao abrir foto do motorista")
		return
	}
	defer openedFile.Close()

	contentType, err := detectImageContentType(openedFile)
	if err != nil {
		respondError(c, http.StatusBadRequest, "Arquivo enviado nao e uma imagem valida", err)
		return
	}

	photoURL, err := h.photoStorage.UploadMotoristaPhoto(c.Request.Context(), openedFile, file.Filename, contentType)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao enviar foto do motorista")
		return
	}

	item, err := h.repo.UpdatePhoto(c.Request.Context(), c.Param("id"), photoURL)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao vincular foto ao motorista")
		return
	}

	respondSuccess(c, http.StatusOK, "Foto do motorista enviada com sucesso", item)
}

func (h *MotoristaHandler) Indicators(c *gin.Context) {
	item, err := h.repo.GetIndicators(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar indicadores do motorista")
		return
	}

	respondSuccess(c, http.StatusOK, "Indicadores do motorista carregados com sucesso", item)
}

func (h *MotoristaHandler) TripsHistory(c *gin.Context) {
	items, err := h.repo.ListTrips(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar historico de viagens do motorista")
		return
	}

	respondSuccess(c, http.StatusOK, "Historico de viagens do motorista carregado com sucesso", items)
}

func (h *MotoristaHandler) OccurrencesHistory(c *gin.Context) {
	items, err := h.repo.ListOccurrences(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar ocorrencias do motorista")
		return
	}

	respondSuccess(c, http.StatusOK, "Ocorrencias do motorista carregadas com sucesso", items)
}

func (h *MotoristaHandler) ShowSelf(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	item, err := h.repo.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar perfil do motorista")
		return
	}

	respondSuccess(c, http.StatusOK, "Perfil do motorista carregado com sucesso", item)
}

func detectImageContentType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	readBytes, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer[:readBytes])
	if !strings.HasPrefix(contentType, "image/") {
		return "", domain.ErrInvalidInput
	}

	return contentType, nil
}
