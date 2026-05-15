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

type NotificacaoListFilter struct {
	DestinatarioTipo string
	DestinatarioID   string
	Lida             *bool
	Page             int
	Limit            int
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

type PushTokenRegisterRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform"`
	DeviceID string `json:"device_id"`
}

type PushTokenDeleteRequest struct {
	Token string `json:"token" binding:"required"`
}

type PushTokenDetail struct {
	ID         string     `json:"id"`
	ActorType  string     `json:"actor_type"`
	ActorID    string     `json:"actor_id"`
	Token      string     `json:"token"`
	Platform   string     `json:"platform,omitempty"`
	DeviceID   string     `json:"device_id,omitempty"`
	Ativo      bool       `json:"ativo"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}
