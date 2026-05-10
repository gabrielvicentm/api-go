package domain

import "time"

const (
	DestinatarioTipoAdmin     = "admin"
	DestinatarioTipoMotorista = "motorista"

	OrigemTipoMotorista = "motorista"
	OrigemTipoSistema   = "sistema"
)

type NotificacaoCreateRequest struct {
	DestinatarioTipo string `json:"destinatario_tipo"`
	DestinatarioID   string `json:"destinatario_id"`
	OrigemTipo       string `json:"origem_tipo"`
	OrigemID         string `json:"origem_id"`
	Titulo           string `json:"titulo" binding:"required"`
	Mensagem         string `json:"mensagem"`
	ReferenciaTipo   string `json:"referencia_tipo"`
	ReferenciaID     string `json:"referencia_id"`
}

type NotificacaoDetail struct {
	ID               string     `json:"id"`
	DestinatarioTipo string     `json:"destinatario_tipo,omitempty"`
	DestinatarioID   string     `json:"destinatario_id,omitempty"`
	OrigemTipo       string     `json:"origem_tipo,omitempty"`
	OrigemID         string     `json:"origem_id,omitempty"`
	Titulo           string     `json:"titulo"`
	Mensagem         string     `json:"mensagem,omitempty"`
	Lida             bool       `json:"lida"`
	ReferenciaTipo   string     `json:"referencia_tipo,omitempty"`
	ReferenciaID     string     `json:"referencia_id,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
}
