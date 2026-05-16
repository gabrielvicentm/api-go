package domain

import "time"

type HistoricoAlteracaoListFilter struct {
	Search     string
	EntidadeID string
	Acao       string
	Usuario    string
	DataInicio string
	DataFim    string
	Page       int
	Limit      int
}

type HistoricoAlteracaoCampo struct {
	Campo         string `json:"campo,omitempty"`
	ValorAnterior string `json:"valor_anterior,omitempty"`
	ValorNovo     string `json:"valor_novo,omitempty"`
}

type HistoricoAlteracaoItem struct {
	ID          string                    `json:"id"`
	Entidade    string                    `json:"entidade"`
	EntidadeID  string                    `json:"entidade_id"`
	Acao        string                    `json:"acao,omitempty"`
	UsuarioID   string                    `json:"usuario_id,omitempty"`
	UsuarioNome string                    `json:"usuario_nome,omitempty"`
	Origem      string                    `json:"origem,omitempty"`
	Resumo      string                    `json:"resumo,omitempty"`
	Alteracoes  []HistoricoAlteracaoCampo `json:"alteracoes,omitempty"`
	CriadoEm    *time.Time                `json:"criado_em,omitempty"`
}
