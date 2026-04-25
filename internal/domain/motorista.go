package domain

import "time"

type MotoristaListFilter struct {
	Search string
	Status string
	Page   int
	Limit  int
}

type MotoristaCreateRequest struct {
	Nome                string `json:"nome" binding:"required,min=3"`
	CPF                 string `json:"cpf" binding:"required"`
	NumeroCNH           string `json:"numero_cnh" binding:"required"`
	TipoCNH             string `json:"tipo_cnh" binding:"required"`
	ValidadeCNH         string `json:"validade_cnh" binding:"required"`
	Telefone            string `json:"telefone"`
	Email               string `json:"email"`
	EnderecoLogradouro  string `json:"endereco_logradouro"`
	EnderecoNumero      string `json:"endereco_numero"`
	EnderecoComplemento string `json:"endereco_complemento"`
	EnderecoBairro      string `json:"endereco_bairro"`
	EnderecoCidade      string `json:"endereco_cidade"`
	EnderecoUF          string `json:"endereco_uf"`
	EnderecoCEP         string `json:"endereco_cep"`
	DataAdmissao        string `json:"data_admissao"`
	Status              string `json:"status"`
	Observacoes         string `json:"observacoes"`
	SenhaInicial        string `json:"senha_inicial" binding:"required,min=6"`
}

type MotoristaUpdateRequest struct {
	Nome                string `json:"nome" binding:"required,min=3"`
	CPF                 string `json:"cpf" binding:"required"`
	NumeroCNH           string `json:"numero_cnh" binding:"required"`
	TipoCNH             string `json:"tipo_cnh" binding:"required"`
	ValidadeCNH         string `json:"validade_cnh" binding:"required"`
	Telefone            string `json:"telefone"`
	Email               string `json:"email"`
	EnderecoLogradouro  string `json:"endereco_logradouro"`
	EnderecoNumero      string `json:"endereco_numero"`
	EnderecoComplemento string `json:"endereco_complemento"`
	EnderecoBairro      string `json:"endereco_bairro"`
	EnderecoCidade      string `json:"endereco_cidade"`
	EnderecoUF          string `json:"endereco_uf"`
	EnderecoCEP         string `json:"endereco_cep"`
	DataAdmissao        string `json:"data_admissao"`
	Status              string `json:"status"`
	Observacoes         string `json:"observacoes"`
	NovaSenha           string `json:"nova_senha"`
}

type MotoristaStatusUpdateRequest struct {
	Status string `json:"status" binding:"required"`
}

type MotoristaListItem struct {
	ID          string     `json:"id"`
	Nome        string     `json:"nome"`
	CPF         string     `json:"cpf"`
	NumeroCNH   string     `json:"numero_cnh"`
	TipoCNH     string     `json:"tipo_cnh"`
	ValidadeCNH string     `json:"validade_cnh"`
	Telefone    string     `json:"telefone,omitempty"`
	Email       string     `json:"email,omitempty"`
	Status      string     `json:"status"`
	FotoURL     string     `json:"foto_url"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type MotoristaDetail struct {
	ID                  string     `json:"id"`
	Nome                string     `json:"nome"`
	CPF                 string     `json:"cpf"`
	NumeroCNH           string     `json:"numero_cnh"`
	TipoCNH             string     `json:"tipo_cnh"`
	ValidadeCNH         string     `json:"validade_cnh"`
	Telefone            string     `json:"telefone,omitempty"`
	Email               string     `json:"email,omitempty"`
	EnderecoLogradouro  string     `json:"endereco_logradouro,omitempty"`
	EnderecoNumero      string     `json:"endereco_numero,omitempty"`
	EnderecoComplemento string     `json:"endereco_complemento,omitempty"`
	EnderecoBairro      string     `json:"endereco_bairro,omitempty"`
	EnderecoCidade      string     `json:"endereco_cidade,omitempty"`
	EnderecoUF          string     `json:"endereco_uf,omitempty"`
	EnderecoCEP         string     `json:"endereco_cep,omitempty"`
	DataAdmissao        string     `json:"data_admissao,omitempty"`
	Status              string     `json:"status"`
	FotoURL             string     `json:"foto_url"`
	Observacoes         string     `json:"observacoes,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}

type MotoristaIndicators struct {
	MotoristaID      string  `json:"motorista_id"`
	Nome             string  `json:"nome"`
	TotalViagens     int64   `json:"total_viagens"`
	TotalKMRodados   float64 `json:"total_km_rodados"`
	TotalOcorrencias int64   `json:"total_ocorrencias"`
	TotalFreteGerado float64 `json:"total_frete_gerado"`
}

type MotoristaTripSummary struct {
	ID                  string  `json:"id"`
	OrigemCidade        string  `json:"origem_cidade"`
	OrigemUF            string  `json:"origem_uf"`
	DestinoCidade       string  `json:"destino_cidade"`
	DestinoUF           string  `json:"destino_uf"`
	Status              string  `json:"status"`
	DataSaida           string  `json:"data_saida"`
	DataChegadaPrevista string  `json:"data_chegada_prevista,omitempty"`
	ValorFrete          float64 `json:"valor_frete,omitempty"`
}

type MotoristaOccurrenceSummary struct {
	ID           string  `json:"id"`
	Tipo         string  `json:"tipo"`
	Descricao    string  `json:"descricao,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	RegistradoEm string  `json:"registrado_em"`
}
