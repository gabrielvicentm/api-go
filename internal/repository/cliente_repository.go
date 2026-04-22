package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClienteRepository struct {
	db *pgxpool.Pool
}

func NewClienteRepository(db *pgxpool.Pool) *ClienteRepository {
	return &ClienteRepository{db: db}
}

func (r *ClienteRepository) List(ctx context.Context, filter domain.ClienteListFilter) ([]domain.ClienteListItem, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM clientes
		WHERE ($1 = '' OR nome ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%' OR cpf_cnpj ILIKE '%' || $1 || '%')
	`

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, filter.Search).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT id, nome, COALESCE(cpf_cnpj, ''), COALESCE(telefone, ''), COALESCE(email, ''), created_at
		FROM clientes
		WHERE ($1 = '' OR nome ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%' OR cpf_cnpj ILIKE '%' || $1 || '%')
		ORDER BY nome ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, filter.Search, filter.Limit, (filter.Page-1)*filter.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.ClienteListItem, 0)
	for rows.Next() {
		var item domain.ClienteListItem
		if err := rows.Scan(
			&item.ID,
			&item.Nome,
			&item.CPFCNPJ,
			&item.Telefone,
			&item.Email,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *ClienteRepository) GetByID(ctx context.Context, id string) (*domain.ClienteDetail, error) {
	const query = `
		SELECT id, nome, COALESCE(cpf_cnpj, ''), COALESCE(telefone, ''), COALESCE(email, ''), created_at, updated_at
		FROM clientes
		WHERE id = $1
		LIMIT 1
	`

	var item domain.ClienteDetail
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Nome,
		&item.CPFCNPJ,
		&item.Telefone,
		&item.Email,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *ClienteRepository) Create(ctx context.Context, input domain.ClienteCreateRequest) (*domain.ClienteDetail, error) {
	const query = `
		INSERT INTO clientes (nome, cpf_cnpj, telefone, email)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''))
		RETURNING id
	`

	var id string
	err := r.db.QueryRow(
		ctx,
		query,
		strings.TrimSpace(input.Nome),
		strings.TrimSpace(input.CPFCNPJ),
		strings.TrimSpace(input.Telefone),
		strings.TrimSpace(strings.ToLower(input.Email)),
	).Scan(&id)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	return r.GetByID(ctx, id)
}

func (r *ClienteRepository) Update(ctx context.Context, id string, input domain.ClienteUpdateRequest) (*domain.ClienteDetail, error) {
	const query = `
		UPDATE clientes
		SET nome = $2, cpf_cnpj = NULLIF($3, ''), telefone = NULLIF($4, ''), email = NULLIF($5, '')
		WHERE id = $1
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		id,
		strings.TrimSpace(input.Nome),
		strings.TrimSpace(input.CPFCNPJ),
		strings.TrimSpace(input.Telefone),
		strings.TrimSpace(strings.ToLower(input.Email)),
	)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *ClienteRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM clientes WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
