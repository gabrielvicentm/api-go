package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ViagemRepository struct {
	db *pgxpool.Pool
}

func NewViagemRepository(db *pgxpool.Pool) *ViagemRepository {
	return &ViagemRepository{db: db}
}

func (r *ViagemRepository) List(ctx context.Context, filter domain.ViagemListFilter) ([]domain.ViagemDetail, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM viagens v
		JOIN motoristas m ON m.id = v.motorista_id
		JOIN funcionarios f ON f.id = m.id
		JOIN veiculos ve ON ve.id = v.veiculo_id
		LEFT JOIN clientes c ON c.id = v.cliente_id
		LEFT JOIN tipos_carga tc ON tc.id = v.tipo_carga_id
		WHERE (
			$1 = ''
			OR v.origem_cidade ILIKE '%' || $1 || '%'
			OR v.destino_cidade ILIKE '%' || $1 || '%'
			OR v.origem_uf ILIKE '%' || $1 || '%'
			OR v.destino_uf ILIKE '%' || $1 || '%'
			OR f.nome ILIKE '%' || $1 || '%'
			OR ve.placa ILIKE '%' || $1 || '%'
			OR ve.modelo ILIKE '%' || $1 || '%'
			OR COALESCE(c.nome, '') ILIKE '%' || $1 || '%'
			OR COALESCE(tc.nome, '') ILIKE '%' || $1 || '%'
		)
		AND ($2 = '' OR v.status::text = $2)
		AND ($3 = '' OR v.motorista_id::text = $3)
		AND ($4 = '' OR v.veiculo_id::text = $4)
		AND ($5 = '' OR COALESCE(v.cliente_id::text, '') = $5)
		AND ($6 = '' OR v.data_saida >= $6::timestamptz)
		AND ($7 = '' OR v.data_saida <= $7::timestamptz)
		AND (NOT $8 OR v.status <> 'concluida')
	`

	var total int64
	if err := r.db.QueryRow(
		ctx,
		countQuery,
		filter.Search,
		filter.Status,
		filter.MotoristaID,
		filter.VeiculoID,
		filter.ClienteID,
		filter.DataSaidaDe,
		filter.DataSaidaAte,
		filter.ExcludeConcluded,
	).Scan(&total); err != nil {
		return nil, 0, mapDatabaseError(err)
	}

	const query = `
		SELECT
			v.id,
			v.motorista_id,
			COALESCE(f.nome, ''),
			v.veiculo_id,
			COALESCE(ve.placa, ''),
			COALESCE(ve.modelo, ''),
			COALESCE(v.cliente_id::text, ''),
			COALESCE(c.nome, ''),
			v.origem_cidade,
			v.origem_uf,
			v.destino_cidade,
			v.destino_uf,
			v.data_saida,
			v.data_chegada_prevista,
			v.data_chegada_real,
			COALESCE(v.distancia_km::text, ''),
			COALESCE(v.tipo_carga_id::text, ''),
			COALESCE(tc.nome, ''),
			COALESCE(v.peso_carga_kg::text, ''),
			COALESCE(v.valor_frete::text, ''),
			v.km_inicial::text,
			COALESCE(v.km_final::text, ''),
			v.status::text,
			COALESCE(v.observacoes, ''),
			v.created_at,
			v.updated_at
		FROM viagens v
		JOIN motoristas m ON m.id = v.motorista_id
		JOIN funcionarios f ON f.id = m.id
		JOIN veiculos ve ON ve.id = v.veiculo_id
		LEFT JOIN clientes c ON c.id = v.cliente_id
		LEFT JOIN tipos_carga tc ON tc.id = v.tipo_carga_id
		WHERE (
			$1 = ''
			OR v.origem_cidade ILIKE '%' || $1 || '%'
			OR v.destino_cidade ILIKE '%' || $1 || '%'
			OR v.origem_uf ILIKE '%' || $1 || '%'
			OR v.destino_uf ILIKE '%' || $1 || '%'
			OR f.nome ILIKE '%' || $1 || '%'
			OR ve.placa ILIKE '%' || $1 || '%'
			OR ve.modelo ILIKE '%' || $1 || '%'
			OR COALESCE(c.nome, '') ILIKE '%' || $1 || '%'
			OR COALESCE(tc.nome, '') ILIKE '%' || $1 || '%'
		)
		AND ($2 = '' OR v.status::text = $2)
		AND ($3 = '' OR v.motorista_id::text = $3)
		AND ($4 = '' OR v.veiculo_id::text = $4)
		AND ($5 = '' OR COALESCE(v.cliente_id::text, '') = $5)
		AND ($6 = '' OR v.data_saida >= $6::timestamptz)
		AND ($7 = '' OR v.data_saida <= $7::timestamptz)
		AND (NOT $8 OR v.status <> 'concluida')
		ORDER BY v.data_saida DESC
		LIMIT $9 OFFSET $10
	`

	rows, err := r.db.Query(
		ctx,
		query,
		filter.Search,
		filter.Status,
		filter.MotoristaID,
		filter.VeiculoID,
		filter.ClienteID,
		filter.DataSaidaDe,
		filter.DataSaidaAte,
		filter.ExcludeConcluded,
		filter.Limit,
		(filter.Page-1)*filter.Limit,
	)
	if err != nil {
		return nil, 0, mapDatabaseError(err)
	}
	defer rows.Close()

	items := make([]domain.ViagemDetail, 0)
	for rows.Next() {
		var item domain.ViagemDetail
		if err := rows.Scan(
			&item.ID,
			&item.MotoristaID,
			&item.MotoristaNome,
			&item.VeiculoID,
			&item.VeiculoPlaca,
			&item.VeiculoModelo,
			&item.ClienteID,
			&item.ClienteNome,
			&item.OrigemCidade,
			&item.OrigemUF,
			&item.DestinoCidade,
			&item.DestinoUF,
			&item.DataSaida,
			&item.DataChegadaPrevista,
			&item.DataChegadaReal,
			&item.DistanciaKM,
			&item.TipoCargaID,
			&item.TipoCargaNome,
			&item.PesoCargaKG,
			&item.ValorFrete,
			&item.KMInicial,
			&item.KMFinal,
			&item.Status,
			&item.Observacoes,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *ViagemRepository) ListChangeHistory(ctx context.Context, filter domain.HistoricoAlteracaoListFilter) ([]domain.HistoricoAlteracaoItem, int64, error) {
	const actionCase = `
		CASE
			WHEN vh.descricao = 'Viagem criada' THEN 'create'
			WHEN vh.campo_alterado = 'status' OR vh.descricao ILIKE '%status%' OR vh.descricao ILIKE '%finalizada%' THEN 'status'
			WHEN COALESCE(vh.campo_alterado, '') <> '' THEN 'update'
			ELSE 'evento'
		END
	`

	const baseFrom = `
		FROM viagem_historico vh
		JOIN viagens v ON v.id = vh.viagem_id
		JOIN motoristas m ON m.id = v.motorista_id
		JOIN funcionarios fm ON fm.id = m.id
		JOIN veiculos ve ON ve.id = v.veiculo_id
		LEFT JOIN clientes c ON c.id = v.cliente_id
		LEFT JOIN usuarios ua ON vh.usuario_tipo = 'admin' AND ua.id = vh.usuario_id
		LEFT JOIN funcionarios fu ON vh.usuario_tipo = 'motorista' AND fu.id = vh.usuario_id
		WHERE (
			$1 = ''
			OR COALESCE(vh.descricao, '') ILIKE '%' || $1 || '%'
			OR COALESCE(vh.campo_alterado, '') ILIKE '%' || $1 || '%'
			OR vh.viagem_id::text ILIKE '%' || $1 || '%'
			OR fm.nome ILIKE '%' || $1 || '%'
			OR ve.placa ILIKE '%' || $1 || '%'
			OR ve.modelo ILIKE '%' || $1 || '%'
			OR COALESCE(c.nome, '') ILIKE '%' || $1 || '%'
		)
		AND ($2 = '' OR vh.viagem_id::text = $2)
		AND ($3 = '' OR ` + actionCase + ` = $3)
		AND (
			$4 = ''
			OR COALESCE(ua.nome, fu.nome, '') ILIKE '%' || $4 || '%'
			OR COALESCE(ua.email, '') ILIKE '%' || $4 || '%'
			OR vh.usuario_id::text ILIKE '%' || $4 || '%'
		)
		AND ($5 = '' OR vh.created_at >= $5::timestamptz)
		AND ($6 = '' OR vh.created_at < ($6::date + INTERVAL '1 day'))
	`

	countQuery := `SELECT COUNT(*) ` + baseFrom

	var total int64
	if err := r.db.QueryRow(
		ctx,
		countQuery,
		filter.Search,
		filter.EntidadeID,
		filter.Acao,
		filter.Usuario,
		filter.DataInicio,
		filter.DataFim,
	).Scan(&total); err != nil {
		return nil, 0, mapDatabaseError(err)
	}

	query := `
		SELECT
			vh.id,
			vh.viagem_id::text,
			` + actionCase + ` AS acao,
			vh.usuario_id::text,
			COALESCE(ua.nome, fu.nome, ''),
			vh.usuario_tipo,
			COALESCE(vh.descricao, ''),
			COALESCE(vh.campo_alterado, ''),
			COALESCE(vh.valor_anterior, ''),
			COALESCE(vh.valor_novo, ''),
			vh.created_at
	` + baseFrom + `
		ORDER BY vh.created_at DESC
		LIMIT $7 OFFSET $8
	`

	rows, err := r.db.Query(
		ctx,
		query,
		filter.Search,
		filter.EntidadeID,
		filter.Acao,
		filter.Usuario,
		filter.DataInicio,
		filter.DataFim,
		filter.Limit,
		(filter.Page-1)*filter.Limit,
	)
	if err != nil {
		return nil, 0, mapDatabaseError(err)
	}
	defer rows.Close()

	items := make([]domain.HistoricoAlteracaoItem, 0)
	for rows.Next() {
		var item domain.HistoricoAlteracaoItem
		var campoAlterado string
		var valorAnterior string
		var valorNovo string
		if err := rows.Scan(
			&item.ID,
			&item.EntidadeID,
			&item.Acao,
			&item.UsuarioID,
			&item.UsuarioNome,
			&item.Origem,
			&item.Resumo,
			&campoAlterado,
			&valorAnterior,
			&valorNovo,
			&item.CriadoEm,
		); err != nil {
			return nil, 0, err
		}

		item.Entidade = "viagem"
		if campoAlterado != "" || valorAnterior != "" || valorNovo != "" {
			item.Alteracoes = []domain.HistoricoAlteracaoCampo{
				{
					Campo:         campoAlterado,
					ValorAnterior: valorAnterior,
					ValorNovo:     valorNovo,
				},
			}
		}

		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *ViagemRepository) GetByID(ctx context.Context, id string) (*domain.ViagemDetail, error) {
	const query = `
		SELECT
			v.id,
			v.motorista_id,
			COALESCE(f.nome, ''),
			v.veiculo_id,
			COALESCE(ve.placa, ''),
			COALESCE(ve.modelo, ''),
			COALESCE(v.cliente_id::text, ''),
			COALESCE(c.nome, ''),
			v.origem_cidade,
			v.origem_uf,
			v.destino_cidade,
			v.destino_uf,
			v.data_saida,
			v.data_chegada_prevista,
			v.data_chegada_real,
			COALESCE(v.distancia_km::text, ''),
			COALESCE(v.tipo_carga_id::text, ''),
			COALESCE(tc.nome, ''),
			COALESCE(v.peso_carga_kg::text, ''),
			COALESCE(v.valor_frete::text, ''),
			v.km_inicial::text,
			COALESCE(v.km_final::text, ''),
			v.status::text,
			COALESCE(v.observacoes, ''),
			v.created_at,
			v.updated_at
		FROM viagens v
		JOIN motoristas m ON m.id = v.motorista_id
		JOIN funcionarios f ON f.id = m.id
		JOIN veiculos ve ON ve.id = v.veiculo_id
		LEFT JOIN clientes c ON c.id = v.cliente_id
		LEFT JOIN tipos_carga tc ON tc.id = v.tipo_carga_id
		WHERE v.id = $1
		LIMIT 1
	`

	var item domain.ViagemDetail
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.MotoristaID,
		&item.MotoristaNome,
		&item.VeiculoID,
		&item.VeiculoPlaca,
		&item.VeiculoModelo,
		&item.ClienteID,
		&item.ClienteNome,
		&item.OrigemCidade,
		&item.OrigemUF,
		&item.DestinoCidade,
		&item.DestinoUF,
		&item.DataSaida,
		&item.DataChegadaPrevista,
		&item.DataChegadaReal,
		&item.DistanciaKM,
		&item.TipoCargaID,
		&item.TipoCargaNome,
		&item.PesoCargaKG,
		&item.ValorFrete,
		&item.KMInicial,
		&item.KMFinal,
		&item.Status,
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

	return &item, nil
}

func (r *ViagemRepository) ListDocuments(ctx context.Context, viagemID string) ([]domain.ViagemDocumentoItem, error) {
	const query = `
		SELECT
			id,
			viagem_id,
			nome,
			tipo,
			url,
			COALESCE(tamanho_bytes, 0),
			created_at
		FROM viagem_documentos
		WHERE viagem_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, viagemID)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	defer rows.Close()

	items := make([]domain.ViagemDocumentoItem, 0)
	for rows.Next() {
		var item domain.ViagemDocumentoItem
		if err := rows.Scan(
			&item.ID,
			&item.ViagemID,
			&item.Nome,
			&item.Tipo,
			&item.URL,
			&item.TamanhoBytes,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *ViagemRepository) GetDocument(ctx context.Context, viagemID, documentID string) (*domain.ViagemDocumentoItem, error) {
	const query = `
		SELECT
			id,
			viagem_id,
			nome,
			tipo,
			url,
			COALESCE(tamanho_bytes, 0),
			created_at
		FROM viagem_documentos
		WHERE viagem_id = $1
		AND id = $2
		LIMIT 1
	`

	var item domain.ViagemDocumentoItem
	err := r.db.QueryRow(ctx, query, strings.TrimSpace(viagemID), strings.TrimSpace(documentID)).Scan(
		&item.ID,
		&item.ViagemID,
		&item.Nome,
		&item.Tipo,
		&item.URL,
		&item.TamanhoBytes,
		&item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, mapDatabaseError(err)
	}

	return &item, nil
}

func (r *ViagemRepository) ListHistory(ctx context.Context, viagemID string) ([]domain.ViagemHistoricoItem, error) {
	const query = `
		SELECT
			id,
			viagem_id,
			usuario_tipo,
			usuario_id,
			COALESCE(campo_alterado, ''),
			COALESCE(valor_anterior, ''),
			COALESCE(valor_novo, ''),
			COALESCE(descricao, ''),
			created_at
		FROM viagem_historico
		WHERE viagem_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, viagemID)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	defer rows.Close()

	items := make([]domain.ViagemHistoricoItem, 0)
	for rows.Next() {
		var item domain.ViagemHistoricoItem
		if err := rows.Scan(
			&item.ID,
			&item.ViagemID,
			&item.UsuarioTipo,
			&item.UsuarioID,
			&item.CampoAlterado,
			&item.ValorAnterior,
			&item.ValorNovo,
			&item.Descricao,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *ViagemRepository) ListStops(ctx context.Context, viagemID string) ([]domain.ViagemParadaItem, error) {
	const query = `
		SELECT
			id,
			viagem_id,
			motivo,
			COALESCE(latitude::text, ''),
			COALESCE(longitude::text, ''),
			iniciada_em,
			finalizada_em
		FROM viagem_paradas
		WHERE viagem_id = $1
		ORDER BY iniciada_em DESC
	`

	rows, err := r.db.Query(ctx, query, viagemID)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	defer rows.Close()

	items := make([]domain.ViagemParadaItem, 0)
	for rows.Next() {
		var item domain.ViagemParadaItem
		if err := rows.Scan(
			&item.ID,
			&item.ViagemID,
			&item.Motivo,
			&item.Latitude,
			&item.Longitude,
			&item.IniciadaEm,
			&item.FinalizadaEm,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *ViagemRepository) StartStop(ctx context.Context, viagemID string, input domain.ViagemParadaStartRequest) (*domain.ViagemParadaItem, *domain.ViagemDetail, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	const currentTripQuery = `
		SELECT status::text
		FROM viagens
		WHERE id = $1
		FOR UPDATE
	`

	var status string
	if err := tx.QueryRow(ctx, currentTripQuery, strings.TrimSpace(viagemID)).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}
		return nil, nil, mapDatabaseError(err)
	}

	if status != "em_andamento" {
		return nil, nil, domain.ErrInvalidInput
	}

	const openStopQuery = `
		SELECT EXISTS (
			SELECT 1
			FROM viagem_paradas
			WHERE viagem_id = $1
			AND finalizada_em IS NULL
		)
	`

	var hasOpenStop bool
	if err := tx.QueryRow(ctx, openStopQuery, strings.TrimSpace(viagemID)).Scan(&hasOpenStop); err != nil {
		return nil, nil, mapDatabaseError(err)
	}
	if hasOpenStop {
		return nil, nil, domain.ErrConflict
	}

	const insertStopQuery = `
		INSERT INTO viagem_paradas (
			viagem_id,
			motivo,
			latitude,
			longitude
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, '')::numeric,
			NULLIF($4, '')::numeric
		)
		RETURNING id, viagem_id, motivo, COALESCE(latitude::text, ''), COALESCE(longitude::text, ''), iniciada_em, finalizada_em
	`

	var item domain.ViagemParadaItem
	if err := tx.QueryRow(
		ctx,
		insertStopQuery,
		strings.TrimSpace(viagemID),
		strings.TrimSpace(input.Motivo),
		strings.TrimSpace(input.Latitude),
		strings.TrimSpace(input.Longitude),
	).Scan(
		&item.ID,
		&item.ViagemID,
		&item.Motivo,
		&item.Latitude,
		&item.Longitude,
		&item.IniciadaEm,
		&item.FinalizadaEm,
	); err != nil {
		return nil, nil, mapDatabaseError(err)
	}

	const updateTripQuery = `
		UPDATE viagens
		SET status = 'parada', updated_at = NOW()
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, updateTripQuery, strings.TrimSpace(viagemID)); err != nil {
		return nil, nil, mapDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	trip, err := r.GetByID(ctx, viagemID)
	if err != nil {
		return nil, nil, err
	}

	return &item, trip, nil
}

func (r *ViagemRepository) FinishOpenStop(ctx context.Context, viagemID string) (*domain.ViagemParadaItem, *domain.ViagemDetail, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	const currentTripQuery = `
		SELECT status::text
		FROM viagens
		WHERE id = $1
		FOR UPDATE
	`

	var status string
	if err := tx.QueryRow(ctx, currentTripQuery, strings.TrimSpace(viagemID)).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}
		return nil, nil, mapDatabaseError(err)
	}

	if status != "parada" {
		return nil, nil, domain.ErrInvalidInput
	}

	const updateStopQuery = `
		UPDATE viagem_paradas
		SET finalizada_em = NOW()
		WHERE id = (
			SELECT id
			FROM viagem_paradas
			WHERE viagem_id = $1
			AND finalizada_em IS NULL
			ORDER BY iniciada_em DESC
			LIMIT 1
		)
		RETURNING id, viagem_id, motivo, COALESCE(latitude::text, ''), COALESCE(longitude::text, ''), iniciada_em, finalizada_em
	`

	var item domain.ViagemParadaItem
	if err := tx.QueryRow(ctx, updateStopQuery, strings.TrimSpace(viagemID)).Scan(
		&item.ID,
		&item.ViagemID,
		&item.Motivo,
		&item.Latitude,
		&item.Longitude,
		&item.IniciadaEm,
		&item.FinalizadaEm,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}
		return nil, nil, mapDatabaseError(err)
	}

	const updateTripQuery = `
		UPDATE viagens
		SET status = 'em_andamento', updated_at = NOW()
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, updateTripQuery, strings.TrimSpace(viagemID)); err != nil {
		return nil, nil, mapDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	trip, err := r.GetByID(ctx, viagemID)
	if err != nil {
		return nil, nil, err
	}

	return &item, trip, nil
}

func (r *ViagemRepository) ListFinalizations(ctx context.Context, viagemID string) ([]domain.ViagemFinalizacaoItem, error) {
	const query = `
		SELECT
			id,
			viagem_id,
			km_final::text,
			status::text,
			COALESCE(observacao_motorista, ''),
			COALESCE(observacao_admin, ''),
			solicitado_em,
			respondido_em
		FROM viagem_finalizacoes
		WHERE viagem_id = $1
		ORDER BY solicitado_em DESC
	`

	rows, err := r.db.Query(ctx, query, viagemID)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	defer rows.Close()

	items := make([]domain.ViagemFinalizacaoItem, 0)
	for rows.Next() {
		var item domain.ViagemFinalizacaoItem
		if err := rows.Scan(
			&item.ID,
			&item.ViagemID,
			&item.KMFinal,
			&item.Status,
			&item.ObservacaoMotorista,
			&item.ObservacaoAdmin,
			&item.SolicitadoEm,
			&item.RespondidoEm,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *ViagemRepository) Create(ctx context.Context, input domain.ViagemCreateRequest) (*domain.ViagemDetail, error) {
	dataSaida, err := parseRequiredTimestamp(input.DataSaida)
	if err != nil {
		return nil, err
	}

	dataChegadaPrevista, err := parseOptionalTimestamp(input.DataChegadaPrevista)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO viagens (
			motorista_id,
			veiculo_id,
			cliente_id,
			origem_cidade,
			origem_uf,
			destino_cidade,
			destino_uf,
			data_saida,
			data_chegada_prevista,
			distancia_km,
			tipo_carga_id,
			peso_carga_kg,
			valor_frete,
			km_inicial,
			observacoes
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, '')::uuid,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			NULLIF($10, '')::numeric,
			NULLIF($11, '')::uuid,
			NULLIF($12, '')::numeric,
			NULLIF($13, '')::numeric,
			NULLIF($14, '')::numeric,
			NULLIF($15, '')
		)
		RETURNING id
	`

	var id string
	err = r.db.QueryRow(
		ctx,
		query,
		strings.TrimSpace(input.MotoristaID),
		strings.TrimSpace(input.VeiculoID),
		strings.TrimSpace(input.ClienteID),
		strings.TrimSpace(input.OrigemCidade),
		strings.ToUpper(strings.TrimSpace(input.OrigemUF)),
		strings.TrimSpace(input.DestinoCidade),
		strings.ToUpper(strings.TrimSpace(input.DestinoUF)),
		dataSaida,
		dataChegadaPrevista,
		strings.TrimSpace(input.DistanciaKM),
		strings.TrimSpace(input.TipoCargaID),
		strings.TrimSpace(input.PesoCargaKG),
		strings.TrimSpace(input.ValorFrete),
		strings.TrimSpace(input.KMInicial),
		strings.TrimSpace(input.Observacoes),
	).Scan(&id)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	return r.GetByID(ctx, id)
}

func (r *ViagemRepository) CreateDocument(ctx context.Context, input domain.ViagemDocumentoCreateInput) (*domain.ViagemDocumentoItem, error) {
	const query = `
		INSERT INTO viagem_documentos (
			viagem_id,
			nome,
			tipo,
			url,
			tamanho_bytes
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		)
		RETURNING id
	`

	var id string
	err := r.db.QueryRow(
		ctx,
		query,
		strings.TrimSpace(input.ViagemID),
		strings.TrimSpace(input.Nome),
		strings.TrimSpace(input.Tipo),
		strings.TrimSpace(input.URL),
		input.TamanhoBytes,
	).Scan(&id)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	return r.GetDocument(ctx, input.ViagemID, id)
}

func (r *ViagemRepository) FinalizeByAdmin(ctx context.Context, id string, input domain.ViagemFinalizacaoAdminRequest) (*domain.ViagemDetail, error) {
	kmFinal := strings.TrimSpace(input.KMFinal)
	if kmFinal == "" {
		return nil, domain.ErrInvalidInput
	}

	kmFinalValue, err := strconv.ParseFloat(kmFinal, 64)
	if err != nil || math.IsNaN(kmFinalValue) || math.IsInf(kmFinalValue, 0) {
		return nil, domain.ErrInvalidInput
	}

	dataChegadaReal, err := parseOptionalTimestamp(input.DataChegadaReal)
	if err != nil {
		return nil, err
	}
	if dataChegadaReal == nil {
		now := time.Now().UTC()
		dataChegadaReal = &now
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const currentQuery = `
		SELECT veiculo_id::text, km_inicial::text, status::text
		FROM viagens
		WHERE id = $1
		LIMIT 1
	`

	var veiculoID string
	var kmInicialRaw string
	var status string
	if err := tx.QueryRow(ctx, currentQuery, strings.TrimSpace(id)).Scan(&veiculoID, &kmInicialRaw, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, mapDatabaseError(err)
	}

	if status == "cancelada" || status == "concluida" {
		return nil, domain.ErrInvalidInput
	}

	kmInicialValue, err := strconv.ParseFloat(strings.TrimSpace(kmInicialRaw), 64)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	if kmFinalValue < kmInicialValue {
		return nil, domain.ErrInvalidInput
	}

	const updateTripQuery = `
		UPDATE viagens
		SET
			km_final = $2::numeric,
			data_chegada_real = $3,
			status = 'concluida',
			updated_at = NOW()
		WHERE id = $1
	`

	tag, err := tx.Exec(ctx, updateTripQuery, strings.TrimSpace(id), kmFinal, dataChegadaReal)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	const updateVehicleQuery = `
		UPDATE veiculos
		SET
			km_atual = $2::numeric,
			status = 'disponivel',
			updated_at = NOW()
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, updateVehicleQuery, veiculoID, kmFinal); err != nil {
		return nil, mapDatabaseError(err)
	}

	const insertFinalizationQuery = `
		INSERT INTO viagem_finalizacoes (
			viagem_id,
			km_final,
			status,
			observacao_admin,
			solicitado_em,
			respondido_em
		)
		VALUES (
			$1,
			$2::numeric,
			'aprovada',
			NULLIF($3, ''),
			$4,
			$4
		)
	`
	if _, err := tx.Exec(
		ctx,
		insertFinalizationQuery,
		strings.TrimSpace(id),
		kmFinal,
		strings.TrimSpace(input.ObservacaoAdmin),
		dataChegadaReal,
	); err != nil {
		return nil, mapDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *ViagemRepository) Update(ctx context.Context, id string, input domain.ViagemUpdateRequest) (*domain.ViagemDetail, error) {
	dataSaida, err := parseRequiredTimestamp(input.DataSaida)
	if err != nil {
		return nil, err
	}

	dataChegadaPrevista, err := parseOptionalTimestamp(input.DataChegadaPrevista)
	if err != nil {
		return nil, err
	}

	dataChegadaReal, err := parseOptionalTimestamp(input.DataChegadaReal)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE viagens
		SET
			motorista_id = $2,
			veiculo_id = $3,
			cliente_id = NULLIF($4, '')::uuid,
			origem_cidade = $5,
			origem_uf = $6,
			destino_cidade = $7,
			destino_uf = $8,
			data_saida = $9,
			data_chegada_prevista = $10,
			data_chegada_real = $11,
			distancia_km = NULLIF($12, '')::numeric,
			tipo_carga_id = NULLIF($13, '')::uuid,
			peso_carga_kg = NULLIF($14, '')::numeric,
			valor_frete = NULLIF($15, '')::numeric,
			km_inicial = NULLIF($16, '')::numeric,
			km_final = NULLIF($17, '')::numeric,
			status = COALESCE(NULLIF($18, '')::status_viagem, status),
			observacoes = NULLIF($19, '')
		WHERE id = $1
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		id,
		strings.TrimSpace(input.MotoristaID),
		strings.TrimSpace(input.VeiculoID),
		strings.TrimSpace(input.ClienteID),
		strings.TrimSpace(input.OrigemCidade),
		strings.ToUpper(strings.TrimSpace(input.OrigemUF)),
		strings.TrimSpace(input.DestinoCidade),
		strings.ToUpper(strings.TrimSpace(input.DestinoUF)),
		dataSaida,
		dataChegadaPrevista,
		dataChegadaReal,
		strings.TrimSpace(input.DistanciaKM),
		strings.TrimSpace(input.TipoCargaID),
		strings.TrimSpace(input.PesoCargaKG),
		strings.TrimSpace(input.ValorFrete),
		strings.TrimSpace(input.KMInicial),
		strings.TrimSpace(input.KMFinal),
		strings.ToLower(strings.TrimSpace(input.Status)),
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

func (r *ViagemRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return mapDatabaseError(err)
	}
	defer tx.Rollback(ctx)

	const currentQuery = `
		SELECT veiculo_id::text
		FROM viagens
		WHERE id = $1
		LIMIT 1
		FOR UPDATE
	`

	var veiculoID string
	if err := tx.QueryRow(ctx, currentQuery, strings.TrimSpace(id)).Scan(&veiculoID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return mapDatabaseError(err)
	}

	if _, err := tx.Exec(ctx, `UPDATE abastecimentos SET viagem_id = NULL WHERE viagem_id = $1`, strings.TrimSpace(id)); err != nil {
		return mapDatabaseError(err)
	}

	if _, err := tx.Exec(ctx, `UPDATE ocorrencias SET viagem_id = NULL WHERE viagem_id = $1`, strings.TrimSpace(id)); err != nil {
		return mapDatabaseError(err)
	}

	tag, err := tx.Exec(ctx, `DELETE FROM viagens WHERE id = $1`, strings.TrimSpace(id))
	if err != nil {
		return mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	const syncVehicleStatusQuery = `
		UPDATE veiculos ve
		SET status = CASE
			WHEN EXISTS (
				SELECT 1
				FROM viagens v
				WHERE v.veiculo_id = ve.id
				  AND v.status IN ('pendente', 'em_andamento', 'parada')
			) THEN 'em_uso'::status_veiculo
			WHEN ve.status = 'em_uso' THEN 'disponivel'::status_veiculo
			ELSE ve.status
		END,
		updated_at = NOW()
		WHERE ve.id = $1
	`

	if _, err := tx.Exec(ctx, syncVehicleStatusQuery, veiculoID); err != nil {
		return mapDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return mapDatabaseError(err)
	}

	return nil
}

func (r *ViagemRepository) CreateHistory(ctx context.Context, input domain.ViagemHistoricoCreateInput) error {
	const query = `
		INSERT INTO viagem_historico (
			viagem_id,
			usuario_tipo,
			usuario_id,
			campo_alterado,
			valor_anterior,
			valor_novo,
			descricao
		)
		VALUES (
			$1,
			$2,
			$3,
			NULLIF($4, ''),
			NULLIF($5, ''),
			NULLIF($6, ''),
			NULLIF($7, '')
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		strings.TrimSpace(input.ViagemID),
		strings.TrimSpace(input.UsuarioTipo),
		strings.TrimSpace(input.UsuarioID),
		strings.TrimSpace(input.CampoAlterado),
		strings.TrimSpace(input.ValorAnterior),
		strings.TrimSpace(input.ValorNovo),
		strings.TrimSpace(input.Descricao),
	)
	if err != nil {
		return mapDatabaseError(err)
	}

	return nil
}

func (r *ViagemRepository) EnsureMotoristaAtivo(ctx context.Context, motoristaID string) error {
	const query = `
		SELECT f.status::text
		FROM motoristas m
		JOIN funcionarios f ON f.id = m.id
		WHERE m.id = $1
		LIMIT 1
	`

	var status string
	err := r.db.QueryRow(ctx, query, strings.TrimSpace(motoristaID)).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}

	if status != "ativo" {
		return fmt.Errorf("motorista precisa estar ativo: %w", domain.ErrInvalidInput)
	}

	return nil
}

func (r *ViagemRepository) EnsureVeiculoDisponivel(ctx context.Context, veiculoID string) error {
	const query = `
		SELECT status::text
		FROM veiculos
		WHERE id = $1
		LIMIT 1
	`

	var status string
	err := r.db.QueryRow(ctx, query, strings.TrimSpace(veiculoID)).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}

	if status != "disponivel" {
		return fmt.Errorf("veiculo precisa estar disponivel: %w", domain.ErrInvalidInput)
	}

	return nil
}

func (r *ViagemRepository) ValidateKMInicial(ctx context.Context, veiculoID, kmInicial string) error {
	kmInicial = strings.TrimSpace(kmInicial)
	if kmInicial == "" {
		return domain.ErrInvalidInput
	}

	kmInicialValue, err := strconv.ParseFloat(kmInicial, 64)
	if err != nil {
		return domain.ErrInvalidInput
	}

	const query = `
		SELECT km_atual::text
		FROM veiculos
		WHERE id = $1
		LIMIT 1
	`

	var kmAtualRaw string
	err = r.db.QueryRow(ctx, query, strings.TrimSpace(veiculoID)).Scan(&kmAtualRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}

	kmAtualValue, err := strconv.ParseFloat(strings.TrimSpace(kmAtualRaw), 64)
	if err != nil {
		return domain.ErrInvalidInput
	}

	if kmInicialValue < kmAtualValue {
		return fmt.Errorf("km inicial nao pode ser menor que km atual do veiculo: %w", domain.ErrInvalidInput)
	}

	return nil
}

func parseOptionalTimestamp(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	layoutsWithTimezone := []string{
		time.RFC3339Nano,
		time.RFC3339,
	}

	for _, layout := range layoutsWithTimezone {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed, nil
		}
	}

	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		location = time.FixedZone("America/Sao_Paulo", -3*60*60)
	}

	localLayouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
	}

	for _, layout := range localLayouts {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return &parsed, nil
		}
	}

	return nil, domain.ErrInvalidInput
}

func parseRequiredTimestamp(value string) (time.Time, error) {
	parsed, err := parseOptionalTimestamp(value)
	if err != nil {
		return time.Time{}, err
	}
	if parsed == nil {
		return time.Time{}, domain.ErrInvalidInput
	}

	return *parsed, nil
}
