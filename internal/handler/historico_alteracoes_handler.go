package handler

import (
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
	"github.com/gin-gonic/gin"
)

type HistoricoAlteracoesHandler struct {
	repo *repository.ViagemRepository
}

func NewHistoricoAlteracoesHandler(repo *repository.ViagemRepository) *HistoricoAlteracoesHandler {
	return &HistoricoAlteracoesHandler{repo: repo}
}

func (h *HistoricoAlteracoesHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("/historico-alteracoes", h.ListAdmin)
}

func (h *HistoricoAlteracoesHandler) ListAdmin(c *gin.Context) {
	page, limit := parsePagination(c)
	entidade := strings.TrimSpace(strings.ToLower(c.Query("entidade")))

	if entidade != "" && entidade != "viagem" && entidade != "viagens" {
		respondList(c, "Historico de alteracoes carregado com sucesso", []domain.HistoricoAlteracaoItem{}, page, limit, 0)
		return
	}

	items, total, err := h.repo.ListChangeHistory(c.Request.Context(), domain.HistoricoAlteracaoListFilter{
		Search:     strings.TrimSpace(c.Query("search")),
		EntidadeID: strings.TrimSpace(c.Query("entidade_id")),
		Acao:       strings.TrimSpace(strings.ToLower(c.Query("acao"))),
		Usuario:    strings.TrimSpace(c.Query("usuario")),
		DataInicio: strings.TrimSpace(c.Query("data_inicio")),
		DataFim:    strings.TrimSpace(c.Query("data_fim")),
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		respondDomainError(c, err, "Erro interno ao buscar historico de alteracoes")
		return
	}

	respondList(c, "Historico de alteracoes carregado com sucesso", items, page, limit, total)
}
