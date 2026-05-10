package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificacaoRepository struct {
	db *pgxpool.Pool
}

func NewNotificacaoRepository(db *pgxpool.Pool) *NotificacaoRepository {
	return &NotificacaoRepository{db: db}
}

func (r *NotificacaoRepository) Create(ctx context.Context, input domain.NotificacaoCreateRequest) (*domain.NotificacaoDetail, error) {
	const query = `
		INSERT INTO notificacoes (
			destinatario_tipo, destinatario_id, origem_tipo, origem_id, titulo,
			mensagem, referencia_tipo, referencia_id
		)
		VALUES (
			NULLIF($1, ''), NULLIF($2, '')::uuid, NULLIF($3, ''), NULLIF($4, '')::uuid,
			$5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, '')::uuid
		)
		RETURNING id
	`

	var id string
	err := r.db.QueryRow(
		ctx,
		query,
		strings.TrimSpace(strings.ToLower(input.DestinatarioTipo)),
		strings.TrimSpace(input.DestinatarioID),
		strings.TrimSpace(strings.ToLower(input.OrigemTipo)),
		strings.TrimSpace(input.OrigemID),
		strings.TrimSpace(input.Titulo),
		strings.TrimSpace(input.Mensagem),
		strings.TrimSpace(strings.ToLower(input.ReferenciaTipo)),
		strings.TrimSpace(input.ReferenciaID),
	).Scan(&id)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	return r.GetByID(ctx, id)
}

func (r *NotificacaoRepository) GetByID(ctx context.Context, id string) (*domain.NotificacaoDetail, error) {
	const query = `
		SELECT
			id,
			COALESCE(destinatario_tipo, ''),
			COALESCE(destinatario_id::text, ''),
			COALESCE(origem_tipo, ''),
			COALESCE(origem_id::text, ''),
			titulo,
			COALESCE(mensagem, ''),
			lida,
			COALESCE(referencia_tipo, ''),
			COALESCE(referencia_id::text, ''),
			created_at
		FROM notificacoes
		WHERE id = $1
		LIMIT 1
	`

	var item domain.NotificacaoDetail
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.DestinatarioTipo,
		&item.DestinatarioID,
		&item.OrigemTipo,
		&item.OrigemID,
		&item.Titulo,
		&item.Mensagem,
		&item.Lida,
		&item.ReferenciaTipo,
		&item.ReferenciaID,
		&item.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &item, nil
}
