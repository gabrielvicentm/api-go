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
		FROM motoristas m
		JOIN funcionarios f ON f.id = m.id
		WHERE ($1 = '' OR f.nome ILIKE '%' || $1 || '%' OR COALESCE(f.email, '') ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR f.status::text = $2)
	`

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, filter.Search, filter.Status).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT
			m.id,
			f.nome,
			CASE
				WHEN f.cpf IS NULL THEN ''
				ELSE '***.***.***-**'
			END AS cpf,
			CASE
				WHEN m.numero_cnh IS NULL THEN ''
				ELSE '***********'
			END AS numero_cnh,
			m.tipo_cnh::text,
			m.validade_cnh,
			COALESCE(f.telefone, ''),
			COALESCE(f.email, ''),
			f.status::text,
			COALESCE(m.foto_url, f.foto_url, ''),
			m.created_at
		FROM motoristas m
		JOIN funcionarios f ON f.id = m.id
		WHERE ($1 = '' OR f.nome ILIKE '%' || $1 || '%' OR COALESCE(f.email, '') ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR f.status::text = $2)
		ORDER BY f.nome ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, filter.Search, filter.Status, filter.Limit, (filter.Page-1)*filter.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.MotoristaListItem, 0)
	for rows.Next() {
		var item domain.MotoristaListItem
		var validade time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Nome,
			&item.CPF,
			&item.NumeroCNH,
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

		item.ValidadeCNH = validade.Format(dateLayout)
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *MotoristaRepository) GetByID(ctx context.Context, id string) (*domain.MotoristaDetail, error) {
	const query = `
		SELECT
			m.id,
			f.nome,
			f.cpf,
			m.numero_cnh,
			m.tipo_cnh::text,
			m.validade_cnh,
			COALESCE(f.telefone, ''),
			COALESCE(f.email, ''),
			COALESCE(f.endereco, ''),
			COALESCE(f.numero, ''),
			COALESCE(f.complemento, ''),
			COALESCE(f.bairro, ''),
			COALESCE(f.cidade, ''),
			COALESCE(f.estado, ''),
			COALESCE(f.cep, ''),
			f.data_admissao,
			f.status::text,
			COALESCE(m.foto_url, f.foto_url, ''),
			COALESCE(m.observacoes, ''),
			m.created_at,
			m.updated_at
		FROM motoristas m
		JOIN funcionarios f ON f.id = m.id
		WHERE m.id = $1
		LIMIT 1
	`

	var detail domain.MotoristaDetail
	var encryptedCPF []byte
	var encryptedCNH []byte
	var validade time.Time
	var dataAdmissao *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&detail.ID,
		&detail.Nome,
		&encryptedCPF,
		&encryptedCNH,
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

	detail.CPF, err = decryptTextField(ctx, r.db, encryptedCPF, r.encryptionKey)
	if err != nil {
		return nil, err
	}

	detail.NumeroCNH, err = decryptTextField(ctx, r.db, encryptedCNH, r.encryptionKey)
	if err != nil {
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

	status := normalizeMotoristaStatus(input.Status)
	cnh := normalizeDigits(input.NumeroCNH)
	if cnh == "" {
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

	funcionarioRepo := NewFuncionarioRepository(r.db, r.encryptionKey)
	funcionarioInput := domain.FuncionarioCreateRequest{
		Nome:         input.Nome,
		CPF:          input.CPF,
		Telefone:     input.Telefone,
		Email:        input.Email,
		CEP:          input.EnderecoCEP,
		Endereco:     input.EnderecoLogradouro,
		Complemento:  input.EnderecoComplemento,
		Numero:       input.EnderecoNumero,
		Bairro:       input.EnderecoBairro,
		Cidade:       input.EnderecoCidade,
		Estado:       input.EnderecoUF,
		Cargo:        "Motorista",
		Setor:        "Operacao",
		DataAdmissao: input.DataAdmissao,
		Status:       status,
	}

	funcionario, err := funcionarioRepo.createBase(ctx, tx, funcionarioInput)
	if err != nil {
		return nil, err
	}

	const motoristaQuery = `
		INSERT INTO motoristas (id, numero_cnh, numero_cnh_hash, tipo_cnh, validade_cnh, foto_url, observacoes)
		VALUES (
			$1,
			pgp_sym_encrypt($2, $7),
			encode(digest($2, 'sha256'), 'hex'),
			$3::tipo_cnh,
			$4,
			NULLIF($5, ''),
			NULLIF($6, '')
		)
	`

	if _, err := tx.Exec(
		ctx,
		motoristaQuery,
		funcionario.ID,
		cnh,
		strings.ToUpper(strings.TrimSpace(input.TipoCNH)),
		validadeCNH,
		"",
		strings.TrimSpace(input.Observacoes),
		r.encryptionKey,
	); err != nil {
		return nil, mapDatabaseError(err)
	}

	const credentialsQuery = `
		INSERT INTO motorista_credenciais (motorista_id, senha_hash, deve_trocar_senha, ativo)
		VALUES ($1, $2, FALSE, TRUE)
	`
	if _, err := tx.Exec(ctx, credentialsQuery, funcionario.ID, passwordHash); err != nil {
		return nil, mapDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, funcionario.ID)
}

func (r *MotoristaRepository) Update(ctx context.Context, id string, input domain.MotoristaUpdateRequest, passwordHash *string) (*domain.MotoristaDetail, error) {
	validadeCNH, err := parseRequiredDate(input.ValidadeCNH)
	if err != nil {
		return nil, err
	}

	status := normalizeMotoristaStatus(input.Status)
	cnh := normalizeDigits(input.NumeroCNH)
	if cnh == "" {
		return nil, domain.ErrInvalidInput
	}

	if err := r.ensureUniqueCNH(ctx, cnh, id); err != nil {
		return nil, err
	}

	funcionarioRepo := NewFuncionarioRepository(r.db, r.encryptionKey)
	currentFuncionario, err := funcionarioRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	funcionarioInput := domain.FuncionarioUpdateRequest{
		Nome:             input.Nome,
		CPF:              input.CPF,
		RG:               currentFuncionario.RG,
		DataNascimento:   currentFuncionario.DataNascimento,
		Telefone:         input.Telefone,
		Email:            input.Email,
		CEP:              input.EnderecoCEP,
		Endereco:         input.EnderecoLogradouro,
		Complemento:      input.EnderecoComplemento,
		Numero:           input.EnderecoNumero,
		Bairro:           input.EnderecoBairro,
		Cidade:           input.EnderecoCidade,
		Estado:           input.EnderecoUF,
		Cargo:            currentFuncionario.Cargo,
		Setor:            currentFuncionario.Setor,
		TipoContrato:     currentFuncionario.TipoContrato,
		DataAdmissao:     input.DataAdmissao,
		DataDemissao:     currentFuncionario.DataDemissao,
		Status:           status,
		SalarioBase:      currentFuncionario.SalarioBase,
		TipoPagamento:    currentFuncionario.TipoPagamento,
		ValorHoraExtra:   currentFuncionario.ValorHoraExtra,
		AdicionalNoturno: currentFuncionario.AdicionalNoturno,
		ValeAlimentacao:  currentFuncionario.ValeAlimentacao,
		OutrosDescontos:  currentFuncionario.OutrosDescontos,
		Banco:            currentFuncionario.Banco,
		Agencia:          currentFuncionario.Agencia,
		Conta:            currentFuncionario.Conta,
		TipoConta:        currentFuncionario.TipoConta,
		ChavePix:         currentFuncionario.ChavePix,
		HorarioEntrada:   currentFuncionario.HorarioEntrada,
		HorarioSaida:     currentFuncionario.HorarioSaida,
		HorarioAlmoco:    currentFuncionario.HorarioAlmoco,
		HorasExtras:      currentFuncionario.HorasExtras,
		Faltas:           currentFuncionario.Faltas,
		Atestados:        currentFuncionario.Atestados,
		Observacoes:      currentFuncionario.Observacoes,
	}

	if _, err := funcionarioRepo.updateBase(ctx, tx, id, funcionarioInput); err != nil {
		return nil, err
	}

	const query = `
		UPDATE motoristas
		SET
			numero_cnh = pgp_sym_encrypt($2, $6),
			numero_cnh_hash = encode(digest($2, 'sha256'), 'hex'),
			tipo_cnh = $3::tipo_cnh,
			validade_cnh = $4,
			observacoes = NULLIF($5, '')
		WHERE id = $1
	`

	tag, err := tx.Exec(
		ctx,
		query,
		id,
		cnh,
		strings.ToUpper(strings.TrimSpace(input.TipoCNH)),
		validadeCNH,
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
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `DELETE FROM motoristas WHERE id = $1`, id)
	if err != nil {
		return mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if _, err := tx.Exec(ctx, `DELETE FROM funcionarios WHERE id = $1`, id); err != nil {
		return mapDatabaseError(err)
	}

	return tx.Commit(ctx)
}

func (r *MotoristaRepository) UpdateStatus(ctx context.Context, id, status string) (*domain.MotoristaDetail, error) {
	const query = `
		UPDATE funcionarios
		SET status = $2::status_funcionario
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
		WHERE numero_cnh_hash = encode(digest($1, 'sha256'), 'hex')
		LIMIT 1
	`
	args := []any{cnh}

	if excludeID != "" {
		query = `
			SELECT id
			FROM motoristas
			WHERE numero_cnh_hash = encode(digest($1, 'sha256'), 'hex')
			  AND id <> $2
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

func maskCNH(cnh string) string {
	if len(cnh) <= 4 {
		return cnh
	}
	return strings.Repeat("*", len(cnh)-4) + cnh[len(cnh)-4:]
}
