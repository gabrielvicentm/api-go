package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AbastecimentoRepository struct {
	db *pgxpool.Pool
}

func NewAbastecimentoRepository(db *pgxpool.Pool) *AbastecimentoRepository {
	return &AbastecimentoRepository{db: db}
}

func (r *AbastecimentoRepository) List(ctx context.Context, filter domain.AbastecimentoListFilter) ([]domain.AbastecimentoItem, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM abastecimentos a
		WHERE ($1 = '' OR a.veiculo_id::text = $1)
		AND ($2 = '' OR a.motorista_id::text = $2)
	`

	var total int64
	if err := r.db.QueryRow(
		ctx,
		countQuery,
		strings.TrimSpace(filter.VeiculoID),
		strings.TrimSpace(filter.MotoristaID),
	).Scan(&total); err != nil {
		return nil, 0, mapDatabaseError(err)
	}

	const query = `
		SELECT
			a.id,
			COALESCE(a.viagem_id::text, ''),
			a.veiculo_id,
			COALESCE(v.placa, ''),
			COALESCE(v.modelo, ''),
			a.motorista_id,
			COALESCE(f.nome, ''),
			a.tipo_combustivel::text,
			a.km_atual::text,
			a.litros::text,
			a.valor_por_litro::text,
			a.valor_total::text,
			COALESCE(a.fornecedor, ''),
			COALESCE(a.foto_url, ''),
			a.registrado_em,
			a.created_at
		FROM abastecimentos a
		JOIN veiculos v ON v.id = a.veiculo_id
		JOIN motoristas m ON m.id = a.motorista_id
		JOIN funcionarios f ON f.id = m.id
		WHERE ($1 = '' OR a.veiculo_id::text = $1)
		AND ($2 = '' OR a.motorista_id::text = $2)
		ORDER BY a.registrado_em DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(
		ctx,
		query,
		strings.TrimSpace(filter.VeiculoID),
		strings.TrimSpace(filter.MotoristaID),
		filter.Limit,
		(filter.Page-1)*filter.Limit,
	)
	if err != nil {
		return nil, 0, mapDatabaseError(err)
	}
	defer rows.Close()

	items := make([]domain.AbastecimentoItem, 0)
	for rows.Next() {
		item, err := scanAbastecimento(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *AbastecimentoRepository) GetByID(ctx context.Context, id string) (*domain.AbastecimentoItem, error) {
	const query = `
		SELECT
			a.id,
			COALESCE(a.viagem_id::text, ''),
			a.veiculo_id,
			COALESCE(v.placa, ''),
			COALESCE(v.modelo, ''),
			a.motorista_id,
			COALESCE(f.nome, ''),
			a.tipo_combustivel::text,
			a.km_atual::text,
			a.litros::text,
			a.valor_por_litro::text,
			a.valor_total::text,
			COALESCE(a.fornecedor, ''),
			COALESCE(a.foto_url, ''),
			a.registrado_em,
			a.created_at
		FROM abastecimentos a
		JOIN veiculos v ON v.id = a.veiculo_id
		JOIN motoristas m ON m.id = a.motorista_id
		JOIN funcionarios f ON f.id = m.id
		WHERE a.id = $1
		LIMIT 1
	`

	item, err := scanAbastecimento(r.db.QueryRow(ctx, query, strings.TrimSpace(id)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, mapDatabaseError(err)
	}

	return &item, nil
}

func (r *AbastecimentoRepository) Create(ctx context.Context, input domain.AbastecimentoCreateInput) (*domain.AbastecimentoItem, error) {
	kmAtual := strings.TrimSpace(input.KMAtual)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const currentVehicleQuery = `
		SELECT km_atual::text
		FROM veiculos
		WHERE id = $1
		FOR UPDATE
	`

	var currentKMRaw string
	if err := tx.QueryRow(ctx, currentVehicleQuery, strings.TrimSpace(input.VeiculoID)).Scan(&currentKMRaw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, mapDatabaseError(err)
	}

	currentKM, err := strconv.ParseFloat(strings.TrimSpace(currentKMRaw), 64)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	newKM, err := strconv.ParseFloat(kmAtual, 64)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	if newKM < currentKM {
		return nil, domain.ErrInvalidInput
	}

	const insertQuery = `
		INSERT INTO abastecimentos (
			viagem_id,
			veiculo_id,
			motorista_id,
			km_atual,
			litros,
			valor_por_litro,
			fornecedor,
			foto_url
		)
		VALUES (
			NULLIF($1, '')::uuid,
			$2,
			$3,
			$4::numeric,
			$5::numeric,
			$6::numeric,
			NULLIF($7, ''),
			NULLIF($8, '')
		)
		RETURNING id
	`

	var id string
	if err := tx.QueryRow(
		ctx,
		insertQuery,
		strings.TrimSpace(input.ViagemID),
		strings.TrimSpace(input.VeiculoID),
		strings.TrimSpace(input.MotoristaID),
		kmAtual,
		strings.TrimSpace(input.Litros),
		strings.TrimSpace(input.ValorPorLitro),
		strings.TrimSpace(input.Fornecedor),
		strings.TrimSpace(input.FotoURL),
	).Scan(&id); err != nil {
		return nil, mapDatabaseError(err)
	}

	const updateVehicleQuery = `
		UPDATE veiculos
		SET km_atual = $2::numeric, updated_at = NOW()
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, updateVehicleQuery, strings.TrimSpace(input.VeiculoID), kmAtual); err != nil {
		return nil, mapDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

type abastecimentoScanner interface {
	Scan(dest ...any) error
}

func scanAbastecimento(scanner abastecimentoScanner) (domain.AbastecimentoItem, error) {
	var item domain.AbastecimentoItem
	err := scanner.Scan(
		&item.ID,
		&item.ViagemID,
		&item.VeiculoID,
		&item.VeiculoPlaca,
		&item.VeiculoModelo,
		&item.MotoristaID,
		&item.MotoristaNome,
		&item.TipoCombustivel,
		&item.KMAtual,
		&item.Litros,
		&item.ValorPorLitro,
		&item.ValorTotal,
		&item.Fornecedor,
		&item.FotoURL,
		&item.RegistradoEm,
		&item.CreatedAt,
	)
	return item, err
}
