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

func (r *NotificacaoRepository) ListByRecipient(ctx context.Context, filter domain.NotificacaoListFilter) ([]domain.NotificacaoDetail, int64, error) {
	const recipientWhere = `
		(
			$1 = 'admin'
			AND (
				(destinatario_tipo IS NULL AND destinatario_id IS NULL)
				OR (
					destinatario_tipo = 'admin'
					AND (destinatario_id IS NULL OR destinatario_id::text = $2)
				)
			)
		)
		OR (
			$1 = 'motorista'
			AND destinatario_tipo = 'motorista'
			AND destinatario_id::text = $2
		)
	`

	const countQuery = `
		SELECT COUNT(*)
		FROM notificacoes
		WHERE (` + recipientWhere + `)
			AND ($3::boolean IS NULL OR lida = $3)
	`

	var total int64
	if err := r.db.QueryRow(
		ctx,
		countQuery,
		strings.TrimSpace(strings.ToLower(filter.DestinatarioTipo)),
		strings.TrimSpace(filter.DestinatarioID),
		filter.Lida,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

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
		WHERE (` + recipientWhere + `)
			AND ($3::boolean IS NULL OR lida = $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(
		ctx,
		query,
		strings.TrimSpace(strings.ToLower(filter.DestinatarioTipo)),
		strings.TrimSpace(filter.DestinatarioID),
		filter.Lida,
		filter.Limit,
		(filter.Page-1)*filter.Limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.NotificacaoDetail, 0)
	for rows.Next() {
		var item domain.NotificacaoDetail
		if err := rows.Scan(
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
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
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

func (r *NotificacaoRepository) MarkAsReadByRecipient(ctx context.Context, id, destinatarioTipo, destinatarioID string) (*domain.NotificacaoDetail, error) {
	const query = `
		UPDATE notificacoes
		SET lida = TRUE
		WHERE id = $1
			AND (
				(
					$2 = 'admin'
					AND (
						(destinatario_tipo IS NULL AND destinatario_id IS NULL)
						OR (
							destinatario_tipo = 'admin'
							AND (destinatario_id IS NULL OR destinatario_id::text = $3)
						)
					)
				)
				OR (
					$2 = 'motorista'
					AND destinatario_tipo = 'motorista'
					AND destinatario_id::text = $3
				)
			)
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		id,
		strings.TrimSpace(strings.ToLower(destinatarioTipo)),
		strings.TrimSpace(destinatarioID),
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *NotificacaoRepository) UpsertPushToken(ctx context.Context, actorType, actorID string, input domain.PushTokenRegisterRequest) (*domain.PushTokenDetail, error) {
	const query = `
		INSERT INTO push_tokens (
			actor_type, actor_id, token, platform, device_id, ativo, last_seen_at
		)
		VALUES (
			$1, $2::uuid, $3, NULLIF($4, ''), NULLIF($5, ''), TRUE, NOW()
		)
		ON CONFLICT (token)
		DO UPDATE SET
			actor_type = EXCLUDED.actor_type,
			actor_id = EXCLUDED.actor_id,
			platform = EXCLUDED.platform,
			device_id = EXCLUDED.device_id,
			ativo = TRUE,
			last_seen_at = NOW(),
			updated_at = NOW()
		RETURNING
			id,
			actor_type,
			actor_id::text,
			token,
			COALESCE(platform, ''),
			COALESCE(device_id, ''),
			ativo,
			last_seen_at,
			created_at,
			updated_at
	`

	var item domain.PushTokenDetail
	if err := r.db.QueryRow(
		ctx,
		query,
		strings.TrimSpace(strings.ToLower(actorType)),
		strings.TrimSpace(actorID),
		strings.TrimSpace(input.Token),
		strings.TrimSpace(strings.ToLower(input.Platform)),
		strings.TrimSpace(input.DeviceID),
	).Scan(
		&item.ID,
		&item.ActorType,
		&item.ActorID,
		&item.Token,
		&item.Platform,
		&item.DeviceID,
		&item.Ativo,
		&item.LastSeenAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, mapDatabaseError(err)
	}

	return &item, nil
}

func (r *NotificacaoRepository) DeactivatePushToken(ctx context.Context, actorType, actorID, token string) error {
	const query = `
		UPDATE push_tokens
		SET ativo = FALSE, updated_at = NOW()
		WHERE actor_type = $1
			AND actor_id::text = $2
			AND token = $3
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		strings.TrimSpace(strings.ToLower(actorType)),
		strings.TrimSpace(actorID),
		strings.TrimSpace(token),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *NotificacaoRepository) DeactivatePushTokenValue(ctx context.Context, token string) error {
	const query = `
		UPDATE push_tokens
		SET ativo = FALSE, updated_at = NOW()
		WHERE token = $1
	`

	_, err := r.db.Exec(ctx, query, strings.TrimSpace(token))
	return err
}

func (r *NotificacaoRepository) ListPushTokensByRecipient(ctx context.Context, destinatarioTipo, destinatarioID string) ([]string, error) {
	destinatarioTipo = strings.TrimSpace(strings.ToLower(destinatarioTipo))
	destinatarioID = strings.TrimSpace(destinatarioID)

	const query = `
		SELECT token
		FROM push_tokens
		WHERE ativo = TRUE
			AND actor_type = $1
			AND ($2 = '' OR actor_id::text = $2)
		ORDER BY last_seen_at DESC NULLS LAST, updated_at DESC
	`

	rows, err := r.db.Query(ctx, query, destinatarioTipo, destinatarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]string, 0)
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}
