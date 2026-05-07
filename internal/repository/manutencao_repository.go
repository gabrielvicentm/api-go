package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ManutencaoRepository struct {
	db *pgxpool.Pool
}

func NewManutencaoRepository(db *pgxpool.Pool) *ManutencaoRepository {
	return &ManutencaoRepository{db: db}
}

func (r *ManutencaoRepository) List(ctx context.Context, filter domain.ManutencaoListFilter) ([]domain.ManutencaoListItem, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM manutencoes m
		INNER JOIN veiculos v ON v.id = m.veiculo_id
		WHERE ($1 = '' OR m.descricao ILIKE '%' || $1 || '%' OR COALESCE(m.oficina, '') ILIKE '%' || $1 || '%' OR v.placa ILIKE '%' || $1 || '%' OR v.modelo ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR m.status::text = $2)
		  AND ($3 = '' OR m.tipo::text = $3)
		  AND ($4 = '' OR m.veiculo_id::text = $4)
	`

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, filter.Search, filter.Status, filter.Tipo, filter.VeiculoID).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT
			m.id,
			m.veiculo_id,
			v.placa,
			v.modelo,
			m.tipo::text,
			m.status::text,
			m.descricao,
			COALESCE(m.oficina, ''),
			COALESCE(m.km_na_manutencao::text, ''),
			COALESCE(m.km_proxima_manutencao::text, ''),
			COALESCE(m.custo::text, ''),
			m.data_agendada,
			m.data_conclusao,
			m.created_at
		FROM manutencoes m
		INNER JOIN veiculos v ON v.id = m.veiculo_id
		WHERE ($1 = '' OR m.descricao ILIKE '%' || $1 || '%' OR COALESCE(m.oficina, '') ILIKE '%' || $1 || '%' OR v.placa ILIKE '%' || $1 || '%' OR v.modelo ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR m.status::text = $2)
		  AND ($3 = '' OR m.tipo::text = $3)
		  AND ($4 = '' OR m.veiculo_id::text = $4)
		ORDER BY COALESCE(m.data_agendada, m.created_at::date) DESC, m.created_at DESC
		LIMIT $5 OFFSET $6
	`

	rows, err := r.db.Query(ctx, query, filter.Search, filter.Status, filter.Tipo, filter.VeiculoID, filter.Limit, (filter.Page-1)*filter.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.ManutencaoListItem, 0)
	for rows.Next() {
		var item domain.ManutencaoListItem
		var dataAgendada *time.Time
		var dataConclusao *time.Time
		if err := rows.Scan(
			&item.ID,
			&item.VeiculoID,
			&item.VeiculoPlaca,
			&item.VeiculoModelo,
			&item.Tipo,
			&item.Status,
			&item.Descricao,
			&item.Oficina,
			&item.KMNaManutencao,
			&item.KMProximaManutencao,
			&item.Custo,
			&dataAgendada,
			&dataConclusao,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		item.DataAgendada = formatOptionalDate(dataAgendada)
		item.DataConclusao = formatOptionalDate(dataConclusao)
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *ManutencaoRepository) GetByID(ctx context.Context, id string) (*domain.ManutencaoDetail, error) {
	const query = `
		SELECT
			m.id,
			m.veiculo_id,
			v.placa,
			v.modelo,
			m.tipo::text,
			m.status::text,
			m.descricao,
			COALESCE(m.oficina, ''),
			COALESCE(m.km_na_manutencao::text, ''),
			COALESCE(m.km_proxima_manutencao::text, ''),
			m.data_agendada,
			m.data_conclusao,
			COALESCE(m.custo::text, ''),
			COALESCE(m.observacoes, ''),
			m.created_at,
			m.updated_at
		FROM manutencoes m
		INNER JOIN veiculos v ON v.id = m.veiculo_id
		WHERE m.id = $1
		LIMIT 1
	`

	var item domain.ManutencaoDetail
	var dataAgendada *time.Time
	var dataConclusao *time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.VeiculoID,
		&item.VeiculoPlaca,
		&item.VeiculoModelo,
		&item.Tipo,
		&item.Status,
		&item.Descricao,
		&item.Oficina,
		&item.KMNaManutencao,
		&item.KMProximaManutencao,
		&dataAgendada,
		&dataConclusao,
		&item.Custo,
		&item.Observacoes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	item.DataAgendada = formatOptionalDate(dataAgendada)
	item.DataConclusao = formatOptionalDate(dataConclusao)

	return &item, nil
}

func (r *ManutencaoRepository) Create(ctx context.Context, input domain.ManutencaoCreateRequest) (*domain.ManutencaoDetail, error) {
	dataAgendada, err := parseOptionalDate(input.DataAgendada)
	if err != nil {
		return nil, err
	}
	dataConclusao, err := parseOptionalDate(input.DataConclusao)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO manutencoes (
			veiculo_id, tipo, status, descricao, oficina, km_na_manutencao,
			km_proxima_manutencao, data_agendada, data_conclusao, custo, observacoes
		)
		VALUES (
			$1, $2::tipo_manutencao, $3::status_manutencao, $4, NULLIF($5, ''),
			NULLIF($6, '')::numeric, NULLIF($7, '')::numeric, $8, $9, NULLIF($10, '')::numeric, NULLIF($11, '')
		)
		RETURNING id
	`

	var id string
	err = r.db.QueryRow(
		ctx,
		query,
		strings.TrimSpace(input.VeiculoID),
		normalizeMaintenanceType(input.Tipo),
		normalizeMaintenanceStatus(input.Status),
		strings.TrimSpace(input.Descricao),
		strings.TrimSpace(input.Oficina),
		strings.TrimSpace(input.KMNaManutencao),
		strings.TrimSpace(input.KMProximaManutencao),
		dataAgendada,
		dataConclusao,
		strings.TrimSpace(input.Custo),
		strings.TrimSpace(input.Observacoes),
	).Scan(&id)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	return r.GetByID(ctx, id)
}

func (r *ManutencaoRepository) Update(ctx context.Context, id string, input domain.ManutencaoUpdateRequest) (*domain.ManutencaoDetail, error) {
	dataAgendada, err := parseOptionalDate(input.DataAgendada)
	if err != nil {
		return nil, err
	}
	dataConclusao, err := parseOptionalDate(input.DataConclusao)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE manutencoes
		SET
			veiculo_id = $2,
			tipo = $3::tipo_manutencao,
			status = $4::status_manutencao,
			descricao = $5,
			oficina = NULLIF($6, ''),
			km_na_manutencao = NULLIF($7, '')::numeric,
			km_proxima_manutencao = NULLIF($8, '')::numeric,
			data_agendada = $9,
			data_conclusao = $10,
			custo = NULLIF($11, '')::numeric,
			observacoes = NULLIF($12, '')
		WHERE id = $1
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		id,
		strings.TrimSpace(input.VeiculoID),
		normalizeMaintenanceType(input.Tipo),
		normalizeMaintenanceStatus(input.Status),
		strings.TrimSpace(input.Descricao),
		strings.TrimSpace(input.Oficina),
		strings.TrimSpace(input.KMNaManutencao),
		strings.TrimSpace(input.KMProximaManutencao),
		dataAgendada,
		dataConclusao,
		strings.TrimSpace(input.Custo),
		strings.TrimSpace(input.Observacoes),
	)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func normalizeMaintenanceType(tipo string) string {
	tipo = strings.TrimSpace(strings.ToLower(tipo))
	if tipo == "" {
		return "preventiva"
	}
	return tipo
}

func normalizeMaintenanceStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return "agendada"
	}
	return status
}
