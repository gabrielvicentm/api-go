package handler

import (
	"net/http"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gabrielvicentm/api-go.git/internal/service"
	"github.com/gin-gonic/gin"
)

type AbastecimentoHandler struct {
	service *service.AbastecimentoService
}

func NewAbastecimentoHandler(service *service.AbastecimentoService) *AbastecimentoHandler {
	return &AbastecimentoHandler{service: service}
}

func (h *AbastecimentoHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/abastecimentos", h.ListAdmin)
	group.GET("/abastecimentos/:id", h.ShowAdmin)
}

func (h *AbastecimentoHandler) RegisterMotoristaRoutes(group *gin.RouterGroup) {
	group.POST("/abastecimentos", h.Create)
	group.GET("/abastecimentos", h.ListMotorista)
}

func (h *AbastecimentoHandler) ListAdmin(c *gin.Context) {
	page, limit := parsePagination(c)
	items, total, err := h.service.List(c.Request.Context(), domain.AbastecimentoListFilter{
		VeiculoID:   strings.TrimSpace(c.Query("veiculo_id")),
		MotoristaID: strings.TrimSpace(c.Query("motorista_id")),
		Page:        page,
		Limit:       limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar abastecimentos")
		return
	}

	respondList(c, "Abastecimentos listados com sucesso", items, page, limit, total)
}

func (h *AbastecimentoHandler) ShowAdmin(c *gin.Context) {
	item, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar abastecimento")
		return
	}

	respondSuccess(c, http.StatusOK, "Abastecimento carregado com sucesso", item)
}

func (h *AbastecimentoHandler) Create(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	var input domain.AbastecimentoCreateRequest
	if err := c.ShouldBind(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados do abastecimento invalidos", err)
		return
	}

	var photoFileName string
	var photoContentType string
	var photoReader multipartFile

	file, err := c.FormFile("foto")
	if err == nil && file != nil {
		openedFile, err := file.Open()
		if err != nil {
			respondDomainError(c, err, "Erro interno ao abrir foto do abastecimento")
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
		respondDomainError(c, err, "Erro interno ao registrar abastecimento")
		return
	}

	respondSuccess(c, http.StatusCreated, "Abastecimento registrado com sucesso", item)
}

func (h *AbastecimentoHandler) ListMotorista(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	page, limit := parsePagination(c)
	items, total, err := h.service.List(c.Request.Context(), domain.AbastecimentoListFilter{
		MotoristaID: claims.UserID,
		Page:        page,
		Limit:       limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar abastecimentos do motorista")
		return
	}

	respondList(c, "Abastecimentos do motorista listados com sucesso", items, page, limit, total)
}

type multipartFile interface {
	Read(p []byte) (n int, err error)
}
