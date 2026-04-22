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

type VeiculoRepository struct {
	db *pgxpool.Pool
}

func NewVeiculoRepository(db *pgxpool.Pool) *VeiculoRepository {
	return &VeiculoRepository{db: db}
}

func (r *VeiculoRepository) List(ctx context.Context, filter domain.VeiculoListFilter) ([]domain.VeiculoListItem, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM veiculos
		WHERE ($1 = '' OR placa ILIKE '%' || $1 || '%' OR modelo ILIKE '%' || $1 || '%' OR marca ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR status::text = $2)
		  AND ($3 = '' OR tipo::text = $3)
	`

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, filter.Search, filter.Status, filter.Tipo).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT
			id, placa, modelo, marca, ano, tipo::text, status::text,
			km_atual::text, COALESCE(capacidade_carga_kg::text, ''), created_at
		FROM veiculos
		WHERE ($1 = '' OR placa ILIKE '%' || $1 || '%' OR modelo ILIKE '%' || $1 || '%' OR marca ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR status::text = $2)
		  AND ($3 = '' OR tipo::text = $3)
		ORDER BY placa ASC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(ctx, query, filter.Search, filter.Status, filter.Tipo, filter.Limit, (filter.Page-1)*filter.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.VeiculoListItem, 0)
	for rows.Next() {
		var item domain.VeiculoListItem
		if err := rows.Scan(
			&item.ID,
			&item.Placa,
			&item.Modelo,
			&item.Marca,
			&item.Ano,
			&item.Tipo,
			&item.Status,
			&item.KMAtual,
			&item.CapacidadeCargaKG,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *VeiculoRepository) GetByID(ctx context.Context, id string) (*domain.VeiculoDetail, error) {
	const query = `
		SELECT
			id, placa, modelo, marca, ano, tipo::text,
			COALESCE(capacidade_carga_kg::text, ''),
			COALESCE(renavam, ''),
			km_atual::text,
			status::text,
			vencimento_seguro,
			vencimento_licenciamento,
			vencimento_ipva,
			COALESCE(seguradora, ''),
			COALESCE(numero_apolice, ''),
			COALESCE(observacoes, ''),
			created_at,
			updated_at
		FROM veiculos
		WHERE id = $1
		LIMIT 1
	`

	var detail domain.VeiculoDetail
	var vencimentoSeguro *time.Time
	var vencimentoLicenciamento *time.Time
	var vencimentoIPVA *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&detail.ID,
		&detail.Placa,
		&detail.Modelo,
		&detail.Marca,
		&detail.Ano,
		&detail.Tipo,
		&detail.CapacidadeCargaKG,
		&detail.Renavam,
		&detail.KMAtual,
		&detail.Status,
		&vencimentoSeguro,
		&vencimentoLicenciamento,
		&vencimentoIPVA,
		&detail.Seguradora,
		&detail.NumeroApolice,
		&detail.Observacoes,
		&detail.CreatedAt,
		&detail.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	detail.VencimentoSeguro = formatOptionalDate(vencimentoSeguro)
	detail.VencimentoLicenciamento = formatOptionalDate(vencimentoLicenciamento)
	detail.VencimentoIPVA = formatOptionalDate(vencimentoIPVA)

	return &detail, nil
}

func (r *VeiculoRepository) Create(ctx context.Context, input domain.VeiculoCreateRequest) (*domain.VeiculoDetail, error) {
	vencimentoSeguro, err := parseOptionalDate(input.VencimentoSeguro)
	if err != nil {
		return nil, err
	}
	vencimentoLicenciamento, err := parseOptionalDate(input.VencimentoLicenciamento)
	if err != nil {
		return nil, err
	}
	vencimentoIPVA, err := parseOptionalDate(input.VencimentoIPVA)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO veiculos (
			placa, modelo, marca, ano, tipo, capacidade_carga_kg, renavam, km_atual, status,
			vencimento_seguro, vencimento_licenciamento, vencimento_ipva, seguradora, numero_apolice, observacoes
		)
		VALUES (
			$1, $2, $3, $4, $5::tipo_veiculo, NULLIF($6, '')::numeric, NULLIF($7, ''),
			COALESCE(NULLIF($8, '')::numeric, 0), $9::status_veiculo, $10, $11, $12,
			NULLIF($13, ''), NULLIF($14, ''), NULLIF($15, '')
		)
		RETURNING id
	`

	var id string
	err = r.db.QueryRow(
		ctx,
		query,
		normalizePlate(input.Placa),
		strings.TrimSpace(input.Modelo),
		strings.TrimSpace(input.Marca),
		input.Ano,
		normalizeVehicleType(input.Tipo),
		strings.TrimSpace(input.CapacidadeCargaKG),
		strings.TrimSpace(input.Renavam),
		strings.TrimSpace(input.KMAtual),
		normalizeVehicleStatus(input.Status),
		vencimentoSeguro,
		vencimentoLicenciamento,
		vencimentoIPVA,
		strings.TrimSpace(input.Seguradora),
		strings.TrimSpace(input.NumeroApolice),
		strings.TrimSpace(input.Observacoes),
	).Scan(&id)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	return r.GetByID(ctx, id)
}

func (r *VeiculoRepository) Update(ctx context.Context, id string, input domain.VeiculoUpdateRequest) (*domain.VeiculoDetail, error) {
	vencimentoSeguro, err := parseOptionalDate(input.VencimentoSeguro)
	if err != nil {
		return nil, err
	}
	vencimentoLicenciamento, err := parseOptionalDate(input.VencimentoLicenciamento)
	if err != nil {
		return nil, err
	}
	vencimentoIPVA, err := parseOptionalDate(input.VencimentoIPVA)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE veiculos
		SET
			placa = $2,
			modelo = $3,
			marca = $4,
			ano = $5,
			tipo = $6::tipo_veiculo,
			capacidade_carga_kg = NULLIF($7, '')::numeric,
			renavam = NULLIF($8, ''),
			km_atual = COALESCE(NULLIF($9, '')::numeric, 0),
			status = $10::status_veiculo,
			vencimento_seguro = $11,
			vencimento_licenciamento = $12,
			vencimento_ipva = $13,
			seguradora = NULLIF($14, ''),
			numero_apolice = NULLIF($15, ''),
			observacoes = NULLIF($16, '')
		WHERE id = $1
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		id,
		normalizePlate(input.Placa),
		strings.TrimSpace(input.Modelo),
		strings.TrimSpace(input.Marca),
		input.Ano,
		normalizeVehicleType(input.Tipo),
		strings.TrimSpace(input.CapacidadeCargaKG),
		strings.TrimSpace(input.Renavam),
		strings.TrimSpace(input.KMAtual),
		normalizeVehicleStatus(input.Status),
		vencimentoSeguro,
		vencimentoLicenciamento,
		vencimentoIPVA,
		strings.TrimSpace(input.Seguradora),
		strings.TrimSpace(input.NumeroApolice),
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

func (r *VeiculoRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM veiculos WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *VeiculoRepository) GetCosts(ctx context.Context, id string) (*domain.VeiculoCostSummary, error) {
	const query = `
		SELECT veiculo_id, placa, modelo, custo_combustivel, custo_manutencao, custo_total
		FROM vw_custo_total_veiculo
		WHERE veiculo_id = $1
		LIMIT 1
	`

	var item domain.VeiculoCostSummary
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.VeiculoID,
		&item.Placa,
		&item.Modelo,
		&item.CustoCombustivel,
		&item.CustoManutencao,
		&item.CustoTotal,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *VeiculoRepository) GetConsumption(ctx context.Context, id string) (*domain.VeiculoConsumptionSummary, error) {
	const query = `
		SELECT veiculo_id, placa, modelo, total_abastecimentos, COALESCE(total_litros, 0),
		       COALESCE(km_percorridos, 0), COALESCE(consumo_km_por_litro, 0), COALESCE(custo_combustivel, 0)
		FROM vw_consumo_veiculo
		WHERE veiculo_id = $1
		LIMIT 1
	`

	var item domain.VeiculoConsumptionSummary
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.VeiculoID,
		&item.Placa,
		&item.Modelo,
		&item.TotalAbastecimentos,
		&item.TotalLitros,
		&item.KMPercorridos,
		&item.ConsumoKMPorLitro,
		&item.CustoCombustivel,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *VeiculoRepository) GetHistory(ctx context.Context, id string) ([]domain.VeiculoHistoryItem, error) {
	const query = `
		SELECT tipo, id, titulo, descricao, data_evento, status
		FROM (
			SELECT
				'viagem' AS tipo,
				v.id::text AS id,
				v.origem_cidade || '/' || v.origem_uf || ' -> ' || v.destino_cidade || '/' || v.destino_uf AS titulo,
				COALESCE(v.observacoes, '') AS descricao,
				v.data_saida AS data_evento,
				v.status::text AS status
			FROM viagens v
			WHERE v.veiculo_id = $1

			UNION ALL

			SELECT
				'abastecimento',
				a.id::text,
				'Abastecimento - ' || COALESCE(a.fornecedor, 'sem fornecedor'),
				COALESCE(a.tipo_combustivel::text, ''),
				a.registrado_em,
				NULL::text
			FROM abastecimentos a
			WHERE a.veiculo_id = $1

			UNION ALL

			SELECT
				'manutencao',
				m.id::text,
				'Manutencao ' || m.tipo::text,
				m.descricao,
				COALESCE(m.updated_at, m.created_at) AS data_evento,
				m.status::text
			FROM manutencoes m
			WHERE m.veiculo_id = $1
		) historico
		ORDER BY data_evento DESC
		LIMIT 100
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.VeiculoHistoryItem, 0)
	for rows.Next() {
		var item domain.VeiculoHistoryItem
		var dataEvento time.Time
		if err := rows.Scan(
			&item.Tipo,
			&item.ID,
			&item.Titulo,
			&item.Descricao,
			&dataEvento,
			&item.Status,
		); err != nil {
			return nil, err
		}

		item.DataEvento = dataEvento.Format(time.RFC3339)
		items = append(items, item)
	}

	return items, rows.Err()
}

func normalizePlate(placa string) string {
	return strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(placa, "-", "")))
}

func normalizeVehicleType(tipo string) string {
	tipo = strings.TrimSpace(strings.ToLower(tipo))
	if tipo == "" {
		return "outro"
	}
	return tipo
}

func normalizeVehicleStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return "disponivel"
	}
	return status
}
