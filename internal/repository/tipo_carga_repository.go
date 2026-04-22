package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TipoCargaRepository struct {
	db *pgxpool.Pool
}

func NewTipoCargaRepository(db *pgxpool.Pool) *TipoCargaRepository {
	return &TipoCargaRepository{db: db}
}

func (r *TipoCargaRepository) List(ctx context.Context, filter domain.TipoCargaListFilter) ([]domain.TipoCargaItem, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM tipos_carga
		WHERE ($1 = '' OR nome ILIKE '%' || $1 || '%' OR descricao ILIKE '%' || $1 || '%')
	`

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, filter.Search).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT id, nome, COALESCE(descricao, ''), created_at
		FROM tipos_carga
		WHERE ($1 = '' OR nome ILIKE '%' || $1 || '%' OR descricao ILIKE '%' || $1 || '%')
		ORDER BY nome ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, filter.Search, filter.Limit, (filter.Page-1)*filter.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.TipoCargaItem, 0)
	for rows.Next() {
		var item domain.TipoCargaItem
		if err := rows.Scan(&item.ID, &item.Nome, &item.Descricao, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *TipoCargaRepository) GetByID(ctx context.Context, id string) (*domain.TipoCargaItem, error) {
	const query = `
		SELECT id, nome, COALESCE(descricao, ''), created_at
		FROM tipos_carga
		WHERE id = $1
		LIMIT 1
	`

	var item domain.TipoCargaItem
	if err := r.db.QueryRow(ctx, query, id).Scan(&item.ID, &item.Nome, &item.Descricao, &item.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *TipoCargaRepository) Create(ctx context.Context, input domain.TipoCargaCreateRequest) (*domain.TipoCargaItem, error) {
	const query = `
		INSERT INTO tipos_carga (nome, descricao)
		VALUES ($1, NULLIF($2, ''))
		RETURNING id
	`

	var id string
	err := r.db.QueryRow(ctx, query, strings.TrimSpace(input.Nome), strings.TrimSpace(input.Descricao)).Scan(&id)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	return r.GetByID(ctx, id)
}

func (r *TipoCargaRepository) Update(ctx context.Context, id string, input domain.TipoCargaUpdateRequest) (*domain.TipoCargaItem, error) {
	const query = `
		UPDATE tipos_carga
		SET nome = $2, descricao = NULLIF($3, '')
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, strings.TrimSpace(input.Nome), strings.TrimSpace(input.Descricao))
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *TipoCargaRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM tipos_carga WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
