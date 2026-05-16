package domain

import "time"

type AbastecimentoListFilter struct {
	VeiculoID   string
	MotoristaID string
	Page        int
	Limit       int
}

type AbastecimentoCreateRequest struct {
	KMAtual       string `form:"km_atual" json:"km_atual" binding:"required"`
	Litros        string `form:"litros" json:"litros" binding:"required"`
	ValorPorLitro string `form:"valor_por_litro" json:"valor_por_litro" binding:"required"`
	ValorTotal    string `form:"valor_total" json:"valor_total"`
	Fornecedor    string `form:"fornecedor" json:"fornecedor"`
}

type AbastecimentoCreateInput struct {
	ViagemID      string
	VeiculoID     string
	MotoristaID   string
	KMAtual       string
	Litros        string
	ValorPorLitro string
	Fornecedor    string
	FotoURL       string
}

type AbastecimentoItem struct {
	ID              string     `json:"id"`
	ViagemID        string     `json:"viagem_id,omitempty"`
	VeiculoID       string     `json:"veiculo_id"`
	VeiculoPlaca    string     `json:"veiculo_placa,omitempty"`
	VeiculoModelo   string     `json:"veiculo_modelo,omitempty"`
	MotoristaID     string     `json:"motorista_id"`
	MotoristaNome   string     `json:"motorista_nome,omitempty"`
	TipoCombustivel string     `json:"tipo_combustivel"`
	KMAtual         string     `json:"km_atual"`
	Litros          string     `json:"litros"`
	ValorPorLitro   string     `json:"valor_por_litro"`
	ValorTotal      string     `json:"valor_total"`
	Fornecedor      string     `json:"fornecedor,omitempty"`
	FotoURL         string     `json:"foto_url,omitempty"`
	RegistradoEm    *time.Time `json:"registrado_em,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}
