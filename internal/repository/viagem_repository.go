package repository

import (
	"context"
	"errors"
	"fmt"
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
		FROM viagens
		WHERE (
			$1 = ''
			OR origem_cidade ILIKE '%' || $1 || '%'
			OR destino_cidade ILIKE '%' || $1 || '%'
			OR origem_uf ILIKE '%' || $1 || '%'
			OR destino_uf ILIKE '%' || $1 || '%'
		)
		AND ($2 = '' OR status::text = $2)
		AND ($3 = '' OR motorista_id::text = $3)
		AND ($4 = '' OR veiculo_id::text = $4)
		AND ($5 = '' OR COALESCE(cliente_id::text, '') = $5)
		AND ($6 = '' OR data_saida >= $6::timestamptz)
		AND ($7 = '' OR data_saida <= $7::timestamptz)
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
	).Scan(&total); err != nil {
		return nil, 0, mapDatabaseError(err)
	}

	const query = `
		SELECT
			id,
			motorista_id,
			veiculo_id,
			COALESCE(cliente_id::text, ''),
			origem_cidade,
			origem_uf,
			destino_cidade,
			destino_uf,
			data_saida,
			data_chegada_prevista,
			data_chegada_real,
			COALESCE(distancia_km::text, ''),
			COALESCE(tipo_carga_id::text, ''),
			COALESCE(peso_carga_kg::text, ''),
			COALESCE(valor_frete::text, ''),
			km_inicial::text,
			COALESCE(km_final::text, ''),
			status::text,
			COALESCE(observacoes, ''),
			created_at,
			updated_at
		FROM viagens
		WHERE (
			$1 = ''
			OR origem_cidade ILIKE '%' || $1 || '%'
			OR destino_cidade ILIKE '%' || $1 || '%'
			OR origem_uf ILIKE '%' || $1 || '%'
			OR destino_uf ILIKE '%' || $1 || '%'
		)
		AND ($2 = '' OR status::text = $2)
		AND ($3 = '' OR motorista_id::text = $3)
		AND ($4 = '' OR veiculo_id::text = $4)
		AND ($5 = '' OR COALESCE(cliente_id::text, '') = $5)
		AND ($6 = '' OR data_saida >= $6::timestamptz)
		AND ($7 = '' OR data_saida <= $7::timestamptz)
		ORDER BY data_saida DESC
		LIMIT $8 OFFSET $9
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
			&item.VeiculoID,
			&item.ClienteID,
			&item.OrigemCidade,
			&item.OrigemUF,
			&item.DestinoCidade,
			&item.DestinoUF,
			&item.DataSaida,
			&item.DataChegadaPrevista,
			&item.DataChegadaReal,
			&item.DistanciaKM,
			&item.TipoCargaID,
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

func (r *ViagemRepository) GetByID(ctx context.Context, id string) (*domain.ViagemDetail, error) {
	const query = `
		SELECT
			id,
			motorista_id,
			veiculo_id,
			COALESCE(cliente_id::text, ''),
			origem_cidade,
			origem_uf,
			destino_cidade,
			destino_uf,
			data_saida,
			data_chegada_prevista,
			data_chegada_real,
			COALESCE(distancia_km::text, ''),
			COALESCE(tipo_carga_id::text, ''),
			COALESCE(peso_carga_kg::text, ''),
			COALESCE(valor_frete::text, ''),
			km_inicial::text,
			COALESCE(km_final::text, ''),
			status::text,
			COALESCE(observacoes, ''),
			created_at,
			updated_at
		FROM viagens
		WHERE id = $1
		LIMIT 1
	`

	var item domain.ViagemDetail
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.MotoristaID,
		&item.VeiculoID,
		&item.ClienteID,
		&item.OrigemCidade,
		&item.OrigemUF,
		&item.DestinoCidade,
		&item.DestinoUF,
		&item.DataSaida,
		&item.DataChegadaPrevista,
		&item.DataChegadaReal,
		&item.DistanciaKM,
		&item.TipoCargaID,
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
			descricao,
			COALESCE(latitude::text, ''),
			COALESCE(longitude::text, ''),
			registrado_em
		FROM viagem_paradas
		WHERE viagem_id = $1
		ORDER BY registrado_em DESC
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
			&item.Descricao,
			&item.Latitude,
			&item.Longitude,
			&item.RegistradoEm,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
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
		strings.TrimSpace(input.Status),
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
		SELECT status::text
		FROM motoristas
		WHERE id = $1
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

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
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
