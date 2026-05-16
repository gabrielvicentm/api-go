package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OcorrenciaRepository struct {
	db *pgxpool.Pool
}

func NewOcorrenciaRepository(db *pgxpool.Pool) *OcorrenciaRepository {
	return &OcorrenciaRepository{db: db}
}

func (r *OcorrenciaRepository) List(ctx context.Context, filter domain.OcorrenciaListFilter) ([]domain.OcorrenciaItem, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM ocorrencias o
		WHERE ($1 = '' OR COALESCE(o.viagem_id::text, '') = $1)
		AND ($2 = '' OR COALESCE(o.veiculo_id::text, '') = $2)
		AND ($3 = '' OR o.motorista_id::text = $3)
	`

	var total int64
	if err := r.db.QueryRow(
		ctx,
		countQuery,
		strings.TrimSpace(filter.ViagemID),
		strings.TrimSpace(filter.VeiculoID),
		strings.TrimSpace(filter.MotoristaID),
	).Scan(&total); err != nil {
		return nil, 0, mapDatabaseError(err)
	}

	const query = `
		SELECT
			o.id,
			COALESCE(o.viagem_id::text, ''),
			COALESCE(o.veiculo_id::text, ''),
			COALESCE(v.placa, ''),
			COALESCE(v.modelo, ''),
			o.motorista_id,
			COALESCE(f.nome, ''),
			o.tipo::text,
			o.motivo,
			COALESCE(o.descricao, ''),
			COALESCE(om.url, ''),
			COALESCE(o.latitude::text, ''),
			COALESCE(o.longitude::text, ''),
			o.registrado_em,
			o.created_at
		FROM ocorrencias o
		JOIN motoristas m ON m.id = o.motorista_id
		JOIN funcionarios f ON f.id = m.id
		LEFT JOIN veiculos v ON v.id = o.veiculo_id
		LEFT JOIN LATERAL (
			SELECT url
			FROM ocorrencia_midias
			WHERE ocorrencia_id = o.id
			AND tipo = 'foto'
			ORDER BY created_at DESC
			LIMIT 1
		) om ON TRUE
		WHERE ($1 = '' OR COALESCE(o.viagem_id::text, '') = $1)
		AND ($2 = '' OR COALESCE(o.veiculo_id::text, '') = $2)
		AND ($3 = '' OR o.motorista_id::text = $3)
		ORDER BY o.registrado_em DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(
		ctx,
		query,
		strings.TrimSpace(filter.ViagemID),
		strings.TrimSpace(filter.VeiculoID),
		strings.TrimSpace(filter.MotoristaID),
		filter.Limit,
		(filter.Page-1)*filter.Limit,
	)
	if err != nil {
		return nil, 0, mapDatabaseError(err)
	}
	defer rows.Close()

	items := make([]domain.OcorrenciaItem, 0)
	for rows.Next() {
		item, err := scanOcorrencia(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *OcorrenciaRepository) GetByID(ctx context.Context, id string) (*domain.OcorrenciaItem, error) {
	const query = `
		SELECT
			o.id,
			COALESCE(o.viagem_id::text, ''),
			COALESCE(o.veiculo_id::text, ''),
			COALESCE(v.placa, ''),
			COALESCE(v.modelo, ''),
			o.motorista_id,
			COALESCE(f.nome, ''),
			o.tipo::text,
			o.motivo,
			COALESCE(o.descricao, ''),
			COALESCE(om.url, ''),
			COALESCE(o.latitude::text, ''),
			COALESCE(o.longitude::text, ''),
			o.registrado_em,
			o.created_at
		FROM ocorrencias o
		JOIN motoristas m ON m.id = o.motorista_id
		JOIN funcionarios f ON f.id = m.id
		LEFT JOIN veiculos v ON v.id = o.veiculo_id
		LEFT JOIN LATERAL (
			SELECT url
			FROM ocorrencia_midias
			WHERE ocorrencia_id = o.id
			AND tipo = 'foto'
			ORDER BY created_at DESC
			LIMIT 1
		) om ON TRUE
		WHERE o.id = $1
		LIMIT 1
	`

	item, err := scanOcorrencia(r.db.QueryRow(ctx, query, strings.TrimSpace(id)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, mapDatabaseError(err)
	}

	return &item, nil
}

func (r *OcorrenciaRepository) Create(ctx context.Context, input domain.OcorrenciaCreateInput) (*domain.OcorrenciaItem, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const insertQuery = `
		INSERT INTO ocorrencias (
			viagem_id,
			veiculo_id,
			motorista_id,
			tipo,
			motivo,
			descricao,
			latitude,
			longitude
		)
		VALUES (
			NULLIF($1, '')::uuid,
			NULLIF($2, '')::uuid,
			$3,
			$4::tipo_ocorrencia,
			$5,
			$6,
			NULLIF($7, '')::numeric,
			NULLIF($8, '')::numeric
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
		strings.TrimSpace(input.Tipo),
		strings.TrimSpace(input.Motivo),
		strings.TrimSpace(input.Descricao),
		strings.TrimSpace(input.Latitude),
		strings.TrimSpace(input.Longitude),
	).Scan(&id); err != nil {
		return nil, mapDatabaseError(err)
	}

	if strings.TrimSpace(input.FotoURL) != "" {
		const insertMediaQuery = `
			INSERT INTO ocorrencia_midias (
				ocorrencia_id,
				tipo,
				url
			)
			VALUES ($1, 'foto', $2)
		`
		if _, err := tx.Exec(ctx, insertMediaQuery, id, strings.TrimSpace(input.FotoURL)); err != nil {
			return nil, mapDatabaseError(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

type ocorrenciaScanner interface {
	Scan(dest ...any) error
}

func scanOcorrencia(scanner ocorrenciaScanner) (domain.OcorrenciaItem, error) {
	var item domain.OcorrenciaItem
	err := scanner.Scan(
		&item.ID,
		&item.ViagemID,
		&item.VeiculoID,
		&item.VeiculoPlaca,
		&item.VeiculoModelo,
		&item.MotoristaID,
		&item.MotoristaNome,
		&item.Tipo,
		&item.Motivo,
		&item.Descricao,
		&item.FotoURL,
		&item.Latitude,
		&item.Longitude,
		&item.RegistradoEm,
		&item.CreatedAt,
	)
	if err != nil {
		return item, err
	}

	return item, nil
}
