package domain

import "time"

type ClienteListFilter struct {
	Search string
	Page   int
	Limit  int
}

type ClienteCreateRequest struct {
	Nome     string `json:"nome" binding:"required,min=2"`
	CPFCNPJ  string `json:"cpf_cnpj"`
	Telefone string `json:"telefone"`
	Email    string `json:"email"`
}

type ClienteUpdateRequest = ClienteCreateRequest

type ClienteListItem struct {
	ID        string     `json:"id"`
	Nome      string     `json:"nome"`
	CPFCNPJ   string     `json:"cpf_cnpj,omitempty"`
	Telefone  string     `json:"telefone,omitempty"`
	Email     string     `json:"email,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type ClienteDetail struct {
	ID        string     `json:"id"`
	Nome      string     `json:"nome"`
	CPFCNPJ   string     `json:"cpf_cnpj,omitempty"`
	Telefone  string     `json:"telefone,omitempty"`
	Email     string     `json:"email,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
