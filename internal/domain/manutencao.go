package domain

import "time"

type ManutencaoListFilter struct {
	Search    string
	Status    string
	Tipo      string
	VeiculoID string
	Page      int
	Limit     int
}

type ManutencaoCreateRequest struct {
	VeiculoID           string `json:"veiculo_id" binding:"required"`
	Tipo                string `json:"tipo" binding:"required"`
	Status              string `json:"status"`
	Descricao           string `json:"descricao" binding:"required"`
	Oficina             string `json:"oficina"`
	KMNaManutencao      string `json:"km_na_manutencao"`
	KMProximaManutencao string `json:"km_proxima_manutencao"`
	DataAgendada        string `json:"data_agendada"`
	DataConclusao       string `json:"data_conclusao"`
	Custo               string `json:"custo"`
	Observacoes         string `json:"observacoes"`
}

type ManutencaoUpdateRequest = ManutencaoCreateRequest

type ManutencaoListItem struct {
	ID                  string     `json:"id"`
	VeiculoID           string     `json:"veiculo_id"`
	VeiculoPlaca        string     `json:"veiculo_placa"`
	VeiculoModelo       string     `json:"veiculo_modelo"`
	Tipo                string     `json:"tipo"`
	Status              string     `json:"status"`
	Descricao           string     `json:"descricao"`
	Oficina             string     `json:"oficina,omitempty"`
	KMNaManutencao      string     `json:"km_na_manutencao,omitempty"`
	KMProximaManutencao string     `json:"km_proxima_manutencao,omitempty"`
	Custo               string     `json:"custo,omitempty"`
	DataAgendada        string     `json:"data_agendada,omitempty"`
	DataConclusao       string     `json:"data_conclusao,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
}

type ManutencaoDetail struct {
	ID                  string     `json:"id"`
	VeiculoID           string     `json:"veiculo_id"`
	VeiculoPlaca        string     `json:"veiculo_placa"`
	VeiculoModelo       string     `json:"veiculo_modelo"`
	Tipo                string     `json:"tipo"`
	Status              string     `json:"status"`
	Descricao           string     `json:"descricao"`
	Oficina             string     `json:"oficina,omitempty"`
	KMNaManutencao      string     `json:"km_na_manutencao,omitempty"`
	KMProximaManutencao string     `json:"km_proxima_manutencao,omitempty"`
	DataAgendada        string     `json:"data_agendada,omitempty"`
	DataConclusao       string     `json:"data_conclusao,omitempty"`
	Custo               string     `json:"custo,omitempty"`
	Observacoes         string     `json:"observacoes,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}
