package handler

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/middleware"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gabrielvicentm/api-go.git/internal/service"
	"github.com/gin-gonic/gin"
)

type ViagemHandler struct {
	repo    *repository.ViagemRepository
	service *service.ViagemService
}

func NewViagemHandler(repo *repository.ViagemRepository, service *service.ViagemService) *ViagemHandler {
	return &ViagemHandler{
		repo:    repo,
		service: service,
	}
}

func (h *ViagemHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/viagens", h.ListAdmin)
	group.POST("/viagens", h.Create)
	group.GET("/viagens/:id", h.ShowAdmin)
	group.PUT("/viagens/:id", h.Update)
	group.GET("/viagens/:id/historico", h.History)
	group.GET("/viagens/:id/documentos", h.DocumentsList)
	group.POST("/viagens/:id/documentos", h.DocumentsUpload)
	group.GET("/viagens/:id/documentos/:documentoId", h.DocumentView)
	group.GET("/viagens/:id/finalizacoes", h.FinalizationsListAdmin)
	group.POST("/viagens/:id/finalizacoes/:finalizacaoId/aprovar", h.ApproveFinalization)
	group.POST("/viagens/:id/finalizacoes/:finalizacaoId/rejeitar", h.RejectFinalization)
}

func (h *ViagemHandler) RegisterMotoristaRoutes(group *gin.RouterGroup) {
	group.GET("/viagens", h.ListMotorista)
	group.GET("/viagens/atual", h.Current)
	group.GET("/viagens/historico", h.HistoryMotorista)
	group.POST("/viagens/paradas", h.CreateStop)
	group.POST("/viagens/finalizacao", h.RequestFinalization)
}

func (h *ViagemHandler) ListAdmin(c *gin.Context) {
	page, limit := parsePagination(c)

	items, total, err := h.repo.List(c.Request.Context(), domain.ViagemListFilter{
		Search:       strings.TrimSpace(c.Query("search")),
		Status:       strings.TrimSpace(c.Query("status")),
		MotoristaID:  strings.TrimSpace(c.Query("motorista_id")),
		VeiculoID:    strings.TrimSpace(c.Query("veiculo_id")),
		ClienteID:    strings.TrimSpace(c.Query("cliente_id")),
		DataSaidaDe:  strings.TrimSpace(c.Query("data_saida_de")),
		DataSaidaAte: strings.TrimSpace(c.Query("data_saida_ate")),
		Page:         page,
		Limit:        limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao listar viagens")
		return
	}

	respondList(c, "Viagens listadas com sucesso", items, page, limit, total)
}

func (h *ViagemHandler) Create(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	var input domain.ViagemCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de cadastro invalidos", err)
		return
	}

	item, err := h.service.Create(c.Request.Context(), input, claims.ActorType, claims.UserID)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao cadastrar viagem")
		return
	}

	respondSuccess(c, http.StatusCreated, "Viagem cadastrada com sucesso", item)
}

func (h *ViagemHandler) ShowAdmin(c *gin.Context) {
	item, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar viagem")
		return
	}

	respondSuccess(c, http.StatusOK, "Viagem carregada com sucesso", item)
}

func (h *ViagemHandler) Update(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	var input domain.ViagemUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "Dados de edicao invalidos", err)
		return
	}

	item, err := h.service.Update(c.Request.Context(), c.Param("id"), input, claims.ActorType, claims.UserID)
	if err != nil {
		respondDomainError(c, err, "Erro interno ao atualizar viagem")
		return
	}

	respondSuccess(c, http.StatusOK, "Viagem atualizada com sucesso", item)
}

func (h *ViagemHandler) History(c *gin.Context) {
	items, err := h.repo.ListHistory(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar historico da viagem")
		return
	}

	respondSuccess(c, http.StatusOK, "Historico da viagem carregado com sucesso", items)
}

func (h *ViagemHandler) DocumentsList(c *gin.Context) {
	items, err := h.repo.ListDocuments(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar documentos da viagem")
		return
	}

	respondSuccess(c, http.StatusOK, "Documentos da viagem carregados com sucesso", items)
}

func (h *ViagemHandler) DocumentsUpload(c *gin.Context) {
	claims, ok := middleware.GetAccessClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": domain.ErrInvalidToken.Error()})
		return
	}

	files := uploadedTripDocumentFiles(c)
	if len(files) == 0 {
		respondError(c, http.StatusBadRequest, "Arquivo de documento obrigatorio", http.ErrMissingFile)
		return
	}

	items := make([]domain.ViagemDocumentoItem, 0, len(files))
	for _, file := range files {
		openedFile, err := file.Open()
		if err != nil {
			respondDomainError(c, err, "Erro interno ao abrir documento da viagem")
			return
		}

		documentType, contentType, err := detectTripDocumentType(openedFile, file.Filename)
		if closeErr := openedFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil {
			respondError(c, http.StatusBadRequest, "Documento deve ser um arquivo PDF ou XML valido", err)
			return
		}

		openedFile, err = file.Open()
		if err != nil {
			respondDomainError(c, err, "Erro interno ao abrir documento da viagem")
			return
		}

		item, err := h.service.UploadDocument(
			c.Request.Context(),
			c.Param("id"),
			openedFile,
			file.Filename,
			documentType,
			contentType,
			file.Size,
			claims.ActorType,
			claims.UserID,
		)
		if closeErr := openedFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil {
			respondDomainError(c, err, "Erro interno ao enviar documento da viagem")
			return
		}

		items = append(items, *item)
	}

	respondSuccess(c, http.StatusCreated, "Documento(s) da viagem enviado(s) com sucesso", items)
}

func (h *ViagemHandler) DocumentView(c *gin.Context) {
	item, err := h.repo.GetDocument(c.Request.Context(), c.Param("id"), c.Param("documentoId"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar documento da viagem")
		return
	}

	c.Redirect(http.StatusFound, item.URL)
}

func (h *ViagemHandler) FinalizationsListAdmin(c *gin.Context) {
	items, err := h.repo.ListFinalizations(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar finalizacoes da viagem")
		return
	}

	respondSuccess(c, http.StatusOK, "Finalizacoes da viagem carregadas com sucesso", items)
}

func (h *ViagemHandler) ApproveFinalization(c *gin.Context) {
	respondProtected(c, "admin.viagens.finalizations.approve", "Aprovacao protegida de finalizacao de viagem")
}

func (h *ViagemHandler) RejectFinalization(c *gin.Context) {
	respondProtected(c, "admin.viagens.finalizations.reject", "Rejeicao protegida de finalizacao de viagem")
}

func (h *ViagemHandler) ListMotorista(c *gin.Context) {
	respondProtected(c, "motorista.viagens.list", "Listagem protegida das viagens do motorista")
}

func (h *ViagemHandler) Current(c *gin.Context) {
	respondProtected(c, "motorista.viagens.atual", "Consulta protegida da viagem atual do motorista")
}

func (h *ViagemHandler) HistoryMotorista(c *gin.Context) {
	respondProtected(c, "motorista.viagens.historico", "Historico protegido de viagens anteriores do motorista")
}

func (h *ViagemHandler) CreateStop(c *gin.Context) {
	respondProtected(c, "motorista.viagens.paradas.create", "Registro protegido de parada da viagem")
}

func (h *ViagemHandler) RequestFinalization(c *gin.Context) {
	respondProtected(c, "motorista.viagens.finalizacao.create", "Solicitacao protegida de finalizacao de viagem")
}

func uploadedTripDocumentFiles(c *gin.Context) []*multipart.FileHeader {
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files := make([]*multipart.FileHeader, 0)
		files = append(files, form.File["documentos"]...)
		files = append(files, form.File["documento"]...)
		return files
	}

	file, err := c.FormFile("documento")
	if err == nil {
		return []*multipart.FileHeader{file}
	}

	return nil
}

func detectTripDocumentType(file multipart.File, filename string) (string, string, error) {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(filename)))
	buffer := make([]byte, 512)
	readBytes, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", err
	}

	sample := buffer[:readBytes]
	switch ext {
	case ".pdf":
		if bytes.HasPrefix(sample, []byte("%PDF")) {
			return "pdf", "application/pdf", nil
		}
	case ".xml":
		trimmed := bytes.TrimSpace(sample)
		if bytes.HasPrefix(trimmed, []byte("<")) {
			return "xml", "application/xml", nil
		}
	}

	return "", "", fmt.Errorf("tipo de documento invalido: %w", domain.ErrInvalidInput)
}
