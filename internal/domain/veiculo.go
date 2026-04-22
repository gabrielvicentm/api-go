package domain

import "time"

type VeiculoListFilter struct {
	Search string
	Status string
	Tipo   string
	Page   int
	Limit  int
}

type VeiculoCreateRequest struct {
	Placa                   string `json:"placa" binding:"required"`
	Modelo                  string `json:"modelo" binding:"required"`
	Marca                   string `json:"marca" binding:"required"`
	Ano                     int    `json:"ano" binding:"required"`
	Tipo                    string `json:"tipo" binding:"required"`
	CapacidadeCargaKG       string `json:"capacidade_carga_kg"`
	Renavam                 string `json:"renavam"`
	KMAtual                 string `json:"km_atual"`
	Status                  string `json:"status"`
	VencimentoSeguro        string `json:"vencimento_seguro"`
	VencimentoLicenciamento string `json:"vencimento_licenciamento"`
	VencimentoIPVA          string `json:"vencimento_ipva"`
	Seguradora              string `json:"seguradora"`
	NumeroApolice           string `json:"numero_apolice"`
	Observacoes             string `json:"observacoes"`
}

type VeiculoUpdateRequest = VeiculoCreateRequest

type VeiculoListItem struct {
	ID                string     `json:"id"`
	Placa             string     `json:"placa"`
	Modelo            string     `json:"modelo"`
	Marca             string     `json:"marca"`
	Ano               int        `json:"ano"`
	Tipo              string     `json:"tipo"`
	Status            string     `json:"status"`
	KMAtual           string     `json:"km_atual"`
	CapacidadeCargaKG string     `json:"capacidade_carga_kg,omitempty"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
}

type VeiculoDetail struct {
	ID                      string     `json:"id"`
	Placa                   string     `json:"placa"`
	Modelo                  string     `json:"modelo"`
	Marca                   string     `json:"marca"`
	Ano                     int        `json:"ano"`
	Tipo                    string     `json:"tipo"`
	CapacidadeCargaKG       string     `json:"capacidade_carga_kg,omitempty"`
	Renavam                 string     `json:"renavam,omitempty"`
	KMAtual                 string     `json:"km_atual"`
	Status                  string     `json:"status"`
	VencimentoSeguro        string     `json:"vencimento_seguro,omitempty"`
	VencimentoLicenciamento string     `json:"vencimento_licenciamento,omitempty"`
	VencimentoIPVA          string     `json:"vencimento_ipva,omitempty"`
	Seguradora              string     `json:"seguradora,omitempty"`
	NumeroApolice           string     `json:"numero_apolice,omitempty"`
	Observacoes             string     `json:"observacoes,omitempty"`
	CreatedAt               *time.Time `json:"created_at,omitempty"`
	UpdatedAt               *time.Time `json:"updated_at,omitempty"`
}

type VeiculoCostSummary struct {
	VeiculoID        string  `json:"veiculo_id"`
	Placa            string  `json:"placa"`
	Modelo           string  `json:"modelo"`
	CustoCombustivel float64 `json:"custo_combustivel"`
	CustoManutencao  float64 `json:"custo_manutencao"`
	CustoTotal       float64 `json:"custo_total"`
}

type VeiculoConsumptionSummary struct {
	VeiculoID           string  `json:"veiculo_id"`
	Placa               string  `json:"placa"`
	Modelo              string  `json:"modelo"`
	TotalAbastecimentos int64   `json:"total_abastecimentos"`
	TotalLitros         float64 `json:"total_litros"`
	KMPercorridos       float64 `json:"km_percorridos"`
	ConsumoKMPorLitro   float64 `json:"consumo_km_por_litro"`
	CustoCombustivel    float64 `json:"custo_combustivel"`
}

type VeiculoHistoryItem struct {
	Tipo       string `json:"tipo"`
	ID         string `json:"id"`
	Titulo     string `json:"titulo"`
	Descricao  string `json:"descricao,omitempty"`
	DataEvento string `json:"data_evento"`
	Status     string `json:"status,omitempty"`
}
