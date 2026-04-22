package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MotoristaRepository struct {
	db            *pgxpool.Pool
	encryptionKey string
}

func NewMotoristaRepository(db *pgxpool.Pool, encryptionKey string) *MotoristaRepository {
	return &MotoristaRepository{
		db:            db,
		encryptionKey: encryptionKey,
	}
}

func (r *MotoristaRepository) List(ctx context.Context, filter domain.MotoristaListFilter) ([]domain.MotoristaListItem, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM motoristas
		WHERE ($1 = '' OR nome ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR status::text = $2)
	`

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, filter.Search, filter.Status).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT
			id,
			nome,
			pgp_sym_decrypt(cpf, $1)::text AS cpf,
			pgp_sym_decrypt(numero_cnh, $1)::text AS numero_cnh,
			tipo_cnh::text,
			validade_cnh,
			COALESCE(telefone, ''),
			COALESCE(email, ''),
			status::text,
			COALESCE(foto_url, ''),
			created_at
		FROM motoristas
		WHERE ($2 = '' OR nome ILIKE '%' || $2 || '%' OR email ILIKE '%' || $2 || '%')
		  AND ($3 = '' OR status::text = $3)
		ORDER BY nome ASC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(ctx, query, r.encryptionKey, filter.Search, filter.Status, filter.Limit, (filter.Page-1)*filter.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.MotoristaListItem, 0)
	for rows.Next() {
		var item domain.MotoristaListItem
		var cpf string
		var cnh string
		var validade time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Nome,
			&cpf,
			&cnh,
			&item.TipoCNH,
			&validade,
			&item.Telefone,
			&item.Email,
			&item.Status,
			&item.FotoURL,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		item.CPF = maskCPF(cpf)
		item.NumeroCNH = maskCNH(cnh)
		item.ValidadeCNH = validade.Format(dateLayout)
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *MotoristaRepository) GetByID(ctx context.Context, id string) (*domain.MotoristaDetail, error) {
	const query = `
		SELECT
			id,
			nome,
			pgp_sym_decrypt(cpf, $2)::text AS cpf,
			pgp_sym_decrypt(numero_cnh, $2)::text AS numero_cnh,
			tipo_cnh::text,
			validade_cnh,
			COALESCE(telefone, ''),
			COALESCE(email, ''),
			COALESCE(endereco_logradouro, ''),
			COALESCE(endereco_numero, ''),
			COALESCE(endereco_complemento, ''),
			COALESCE(endereco_bairro, ''),
			COALESCE(endereco_cidade, ''),
			COALESCE(endereco_uf, ''),
			COALESCE(endereco_cep, ''),
			data_admissao,
			status::text,
			COALESCE(foto_url, ''),
			COALESCE(observacoes, ''),
			created_at,
			updated_at
		FROM motoristas
		WHERE id = $1
		LIMIT 1
	`

	var detail domain.MotoristaDetail
	var validade time.Time
	var dataAdmissao *time.Time

	err := r.db.QueryRow(ctx, query, id, r.encryptionKey).Scan(
		&detail.ID,
		&detail.Nome,
		&detail.CPF,
		&detail.NumeroCNH,
		&detail.TipoCNH,
		&validade,
		&detail.Telefone,
		&detail.Email,
		&detail.EnderecoLogradouro,
		&detail.EnderecoNumero,
		&detail.EnderecoComplemento,
		&detail.EnderecoBairro,
		&detail.EnderecoCidade,
		&detail.EnderecoUF,
		&detail.EnderecoCEP,
		&dataAdmissao,
		&detail.Status,
		&detail.FotoURL,
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

	detail.ValidadeCNH = validade.Format(dateLayout)
	detail.DataAdmissao = formatOptionalDate(dataAdmissao)

	return &detail, nil
}

func (r *MotoristaRepository) Create(ctx context.Context, input domain.MotoristaCreateRequest, passwordHash string) (*domain.MotoristaDetail, error) {
	validadeCNH, err := parseRequiredDate(input.ValidadeCNH)
	if err != nil {
		return nil, err
	}

	dataAdmissao, err := parseOptionalDate(input.DataAdmissao)
	if err != nil {
		return nil, err
	}

	status := normalizeMotoristaStatus(input.Status)
	cpf := normalizeDigits(input.CPF)
	cnh := normalizeDigits(input.NumeroCNH)

	if len(cpf) != 11 || cnh == "" {
		return nil, domain.ErrInvalidInput
	}

	if err := r.ensureUniqueCNH(ctx, cnh, ""); err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const motoristaQuery = `
		INSERT INTO motoristas (
			nome, cpf, cpf_hash, numero_cnh, tipo_cnh, validade_cnh, telefone, email,
			endereco_logradouro, endereco_numero, endereco_complemento, endereco_bairro,
			endereco_cidade, endereco_uf, endereco_cep, data_admissao, status, observacoes
		)
		VALUES (
			$1, pgp_sym_encrypt($2, $18), encode(digest($2, 'sha256'), 'hex'),
			pgp_sym_encrypt($3, $18), $4::tipo_cnh, $5, NULLIF($6, ''), NULLIF($7, ''),
			NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''),
			NULLIF($12, ''), NULLIF($13, ''), NULLIF($14, ''), $15, $16::status_motorista, NULLIF($17, '')
		)
		RETURNING id
	`

	var id string
	err = tx.QueryRow(
		ctx,
		motoristaQuery,
		strings.TrimSpace(input.Nome),
		cpf,
		cnh,
		strings.ToUpper(strings.TrimSpace(input.TipoCNH)),
		validadeCNH,
		strings.TrimSpace(input.Telefone),
		normalizeNullableEmail(input.Email),
		strings.TrimSpace(input.EnderecoLogradouro),
		strings.TrimSpace(input.EnderecoNumero),
		strings.TrimSpace(input.EnderecoComplemento),
		strings.TrimSpace(input.EnderecoBairro),
		strings.TrimSpace(input.EnderecoCidade),
		strings.ToUpper(strings.TrimSpace(input.EnderecoUF)),
		normalizeDigits(input.EnderecoCEP),
		dataAdmissao,
		status,
		strings.TrimSpace(input.Observacoes),
		r.encryptionKey,
	).Scan(&id)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	const credentialsQuery = `
		INSERT INTO motorista_credenciais (motorista_id, senha_hash, deve_trocar_senha, ativo)
		VALUES ($1, $2, FALSE, TRUE)
	`
	if _, err := tx.Exec(ctx, credentialsQuery, id, passwordHash); err != nil {
		return nil, mapDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *MotoristaRepository) Update(ctx context.Context, id string, input domain.MotoristaUpdateRequest, passwordHash *string) (*domain.MotoristaDetail, error) {
	validadeCNH, err := parseRequiredDate(input.ValidadeCNH)
	if err != nil {
		return nil, err
	}

	dataAdmissao, err := parseOptionalDate(input.DataAdmissao)
	if err != nil {
		return nil, err
	}

	status := normalizeMotoristaStatus(input.Status)
	cpf := normalizeDigits(input.CPF)
	cnh := normalizeDigits(input.NumeroCNH)

	if len(cpf) != 11 || cnh == "" {
		return nil, domain.ErrInvalidInput
	}

	if err := r.ensureUniqueCNH(ctx, cnh, id); err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE motoristas
		SET
			nome = $2,
			cpf = pgp_sym_encrypt($3, $19),
			cpf_hash = encode(digest($3, 'sha256'), 'hex'),
			numero_cnh = pgp_sym_encrypt($4, $19),
			tipo_cnh = $5::tipo_cnh,
			validade_cnh = $6,
			telefone = NULLIF($7, ''),
			email = NULLIF($8, ''),
			endereco_logradouro = NULLIF($9, ''),
			endereco_numero = NULLIF($10, ''),
			endereco_complemento = NULLIF($11, ''),
			endereco_bairro = NULLIF($12, ''),
			endereco_cidade = NULLIF($13, ''),
			endereco_uf = NULLIF($14, ''),
			endereco_cep = NULLIF($15, ''),
			data_admissao = $16,
			status = $17::status_motorista,
			observacoes = NULLIF($18, '')
		WHERE id = $1
	`

	tag, err := tx.Exec(
		ctx,
		query,
		id,
		strings.TrimSpace(input.Nome),
		cpf,
		cnh,
		strings.ToUpper(strings.TrimSpace(input.TipoCNH)),
		validadeCNH,
		strings.TrimSpace(input.Telefone),
		normalizeNullableEmail(input.Email),
		strings.TrimSpace(input.EnderecoLogradouro),
		strings.TrimSpace(input.EnderecoNumero),
		strings.TrimSpace(input.EnderecoComplemento),
		strings.TrimSpace(input.EnderecoBairro),
		strings.TrimSpace(input.EnderecoCidade),
		strings.ToUpper(strings.TrimSpace(input.EnderecoUF)),
		normalizeDigits(input.EnderecoCEP),
		dataAdmissao,
		status,
		strings.TrimSpace(input.Observacoes),
		r.encryptionKey,
	)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	if passwordHash != nil {
		const passwordQuery = `
			UPDATE motorista_credenciais
			SET senha_hash = $2, deve_trocar_senha = FALSE
			WHERE motorista_id = $1
		`
		if _, err := tx.Exec(ctx, passwordQuery, id, *passwordHash); err != nil {
			return nil, mapDatabaseError(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *MotoristaRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM motoristas WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *MotoristaRepository) UpdateStatus(ctx context.Context, id, status string) (*domain.MotoristaDetail, error) {
	const query = `
		UPDATE motoristas
		SET status = $2::status_motorista
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, normalizeMotoristaStatus(status))
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *MotoristaRepository) UpdatePhoto(ctx context.Context, id, photoURL string) (*domain.MotoristaDetail, error) {
	const query = `
		UPDATE motoristas
		SET foto_url = $2
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, photoURL)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *MotoristaRepository) GetIndicators(ctx context.Context, id string) (*domain.MotoristaIndicators, error) {
	const query = `
		SELECT motorista_id, nome, total_viagens, total_km_rodados, total_ocorrencias, total_frete_gerado
		FROM vw_indicadores_motorista
		WHERE motorista_id = $1
		LIMIT 1
	`

	var item domain.MotoristaIndicators
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.MotoristaID,
		&item.Nome,
		&item.TotalViagens,
		&item.TotalKMRodados,
		&item.TotalOcorrencias,
		&item.TotalFreteGerado,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *MotoristaRepository) ListTrips(ctx context.Context, id string) ([]domain.MotoristaTripSummary, error) {
	const query = `
		SELECT
			id,
			origem_cidade,
			origem_uf,
			destino_cidade,
			destino_uf,
			status::text,
			data_saida,
			data_chegada_prevista,
			COALESCE(valor_frete, 0)
		FROM viagens
		WHERE motorista_id = $1
		ORDER BY data_saida DESC
		LIMIT 50
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.MotoristaTripSummary, 0)
	for rows.Next() {
		var item domain.MotoristaTripSummary
		var saida time.Time
		var chegada *time.Time
		if err := rows.Scan(
			&item.ID,
			&item.OrigemCidade,
			&item.OrigemUF,
			&item.DestinoCidade,
			&item.DestinoUF,
			&item.Status,
			&saida,
			&chegada,
			&item.ValorFrete,
		); err != nil {
			return nil, err
		}

		item.DataSaida = saida.Format(time.RFC3339)
		if chegada != nil {
			item.DataChegadaPrevista = chegada.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *MotoristaRepository) ListOccurrences(ctx context.Context, id string) ([]domain.MotoristaOccurrenceSummary, error) {
	const query = `
		SELECT
			id,
			tipo::text,
			COALESCE(descricao, ''),
			COALESCE(latitude, 0),
			COALESCE(longitude, 0),
			registrado_em
		FROM ocorrencias
		WHERE motorista_id = $1
		ORDER BY registrado_em DESC
		LIMIT 50
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.MotoristaOccurrenceSummary, 0)
	for rows.Next() {
		var item domain.MotoristaOccurrenceSummary
		var registradoEm time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Tipo,
			&item.Descricao,
			&item.Latitude,
			&item.Longitude,
			&registradoEm,
		); err != nil {
			return nil, err
		}

		item.RegistradoEm = registradoEm.Format(time.RFC3339)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *MotoristaRepository) ensureUniqueCNH(ctx context.Context, cnh, excludeID string) error {
	query := `
		SELECT id
		FROM motoristas
		WHERE pgp_sym_decrypt(numero_cnh, $2)::text = $1
		LIMIT 1
	`
	args := []any{cnh, r.encryptionKey}

	if excludeID != "" {
		query = `
			SELECT id
			FROM motoristas
			WHERE pgp_sym_decrypt(numero_cnh, $2)::text = $1
			  AND id <> $3
			LIMIT 1
		`
		args = append(args, excludeID)
	}

	var foundID string
	err := r.db.QueryRow(ctx, query, args...).Scan(&foundID)
	if err == nil {
		return domain.ErrConflict
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	return err
}

func normalizeMotoristaStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return "ativo"
	}
	return status
}

func normalizeNullableEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func normalizeDigits(value string) string {
	replacer := strings.NewReplacer(".", "", "-", "", "/", "", "(", "", ")", "", " ", "")
	return replacer.Replace(strings.TrimSpace(value))
}

func maskCPF(cpf string) string {
	if len(cpf) != 11 {
		return cpf
	}
	return fmt.Sprintf("%s.%s.%s-%s", cpf[:3], cpf[3:6], cpf[6:9], cpf[9:])
}

func maskCNH(cnh string) string {
	if len(cnh) <= 4 {
		return cnh
	}
	return strings.Repeat("*", len(cnh)-4) + cnh[len(cnh)-4:]
}
