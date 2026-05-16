package domain

import "time"

type OcorrenciaListFilter struct {
	Search      string
	Tipo        string
	ViagemID    string
	VeiculoID   string
	MotoristaID string
	Page        int
	Limit       int
}

type OcorrenciaCreateRequest struct {
	Tipo      string `form:"tipo" json:"tipo" binding:"required"`
	Motivo    string `form:"motivo" json:"motivo" binding:"required"`
	Descricao string `form:"descricao" json:"descricao" binding:"required"`
	Latitude  string `form:"latitude" json:"latitude"`
	Longitude string `form:"longitude" json:"longitude"`
}

type OcorrenciaCreateInput struct {
	ViagemID    string
	VeiculoID   string
	MotoristaID string
	Tipo        string
	Motivo      string
	Descricao   string
	Latitude    string
	Longitude   string
	FotoURL     string
}

type OcorrenciaItem struct {
	ID            string     `json:"id"`
	ViagemID      string     `json:"viagem_id,omitempty"`
	VeiculoID     string     `json:"veiculo_id,omitempty"`
	VeiculoPlaca  string     `json:"veiculo_placa,omitempty"`
	VeiculoModelo string     `json:"veiculo_modelo,omitempty"`
	MotoristaID   string     `json:"motorista_id"`
	MotoristaNome string     `json:"motorista_nome,omitempty"`
	Tipo          string     `json:"tipo"`
	Motivo        string     `json:"motivo"`
	Descricao     string     `json:"descricao"`
	FotoURL       string     `json:"foto_url,omitempty"`
	Latitude      string     `json:"latitude,omitempty"`
	Longitude     string     `json:"longitude,omitempty"`
	RegistradoEm  *time.Time `json:"registrado_em,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}
