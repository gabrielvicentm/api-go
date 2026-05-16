package handler

import (
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gabrielvicentm/api-go.git/internal/service"
	"github.com/gin-gonic/gin"
)

type OcorrenciaHandler struct {
	service *service.OcorrenciaService
}

func NewOcorrenciaHandler(service *service.OcorrenciaService) *OcorrenciaHandler {
	return &OcorrenciaHandler{service: service}
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
	page, limit := parsePagination(c)
	items, total, err := h.service.List(c.Request.Context(), domain.OcorrenciaListFilter{
		ViagemID:    strings.TrimSpace(c.Query("viagem_id")),
		VeiculoID:   strings.TrimSpace(c.Query("veiculo_id")),
		MotoristaID: strings.TrimSpace(c.Query("motorista_id")),
		Page:        page,
		Limit:       limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar ocorrencias")
		return
	}

	respondList(c, "Ocorrencias listadas com sucesso", items, page, limit, total)
}

func (h *OcorrenciaHandler) ShowAdmin(c *gin.Context) {
	item, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar ocorrencia")
		return
	}

	respondSuccess(c, http.StatusOK, "Ocorrencia carregada com sucesso", item)
}

func (h *OcorrenciaHandler) Create(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	var input domain.OcorrenciaCreateRequest
	if err := c.ShouldBind(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados da ocorrencia invalidos", err)
		return
	}

	var photoFileName string
	var photoContentType string
	var photoReader multipartFile

	file, err := c.FormFile("foto")
	if err == nil && file != nil {
		openedFile, err := file.Open()
		if err != nil {
			respondDomainError(c, err, "Erro interno ao abrir foto da ocorrencia")
			return
		}
		defer openedFile.Close()

		contentType, err := detectImageContentType(openedFile)
		if err != nil {
			respondError(c, http.StatusBadRequest, "Arquivo enviado nao e uma imagem valida", err)
			return
		}

		photoFileName = file.Filename
		photoContentType = contentType
		photoReader = openedFile
	}

	item, err := h.service.Create(c.Request.Context(), claims.UserID, input, photoReader, photoFileName, photoContentType)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao registrar ocorrencia")
		return
	}

	respondSuccess(c, http.StatusCreated, "Ocorrencia registrada com sucesso", item)
}

func (h *OcorrenciaHandler) ListMotorista(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	page, limit := parsePagination(c)
	items, total, err := h.service.List(c.Request.Context(), domain.OcorrenciaListFilter{
		MotoristaID: claims.UserID,
		Page:        page,
		Limit:       limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar ocorrencias do motorista")
		return
	}

	respondList(c, "Ocorrencias do motorista listadas com sucesso", items, page, limit, total)
}
