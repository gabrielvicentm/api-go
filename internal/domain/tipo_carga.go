package domain

import "time"

type TipoCargaListFilter struct {
	Search string
	Page   int
	Limit  int
}

type TipoCargaCreateRequest struct {
	Nome      string `json:"nome" binding:"required,min=2"`
	Descricao string `json:"descricao"`
}

type TipoCargaUpdateRequest = TipoCargaCreateRequest

type TipoCargaItem struct {
	ID        string     `json:"id"`
	Nome      string     `json:"nome"`
	Descricao string     `json:"descricao,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}
