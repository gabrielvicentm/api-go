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

type FuncionarioRepository struct {
	db            *pgxpool.Pool
	encryptionKey string
}

func NewFuncionarioRepository(db *pgxpool.Pool, encryptionKey string) *FuncionarioRepository {
	return &FuncionarioRepository{
		db:            db,
		encryptionKey: encryptionKey,
	}
}

func (r *FuncionarioRepository) List(ctx context.Context, filter domain.FuncionarioListFilter) ([]domain.FuncionarioListItem, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM funcionarios f
		LEFT JOIN motoristas m ON m.id = f.id
		WHERE ($1 = '' OR f.nome ILIKE '%' || $1 || '%' OR COALESCE(f.email, '') ILIKE '%' || $1 || '%' OR COALESCE(f.cargo, '') ILIKE '%' || $1 || '%' OR COALESCE(f.setor, '') ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR f.status::text = $2)
		  AND ($3 = '' OR CASE WHEN m.id IS NOT NULL THEN 'motorista' ELSE 'funcionario' END = $3)
		  AND ($4 OR m.id IS NULL)
	`

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, filter.Search, filter.Status, normalizeFuncionarioTipo(filter.Tipo), filter.IncludeMotorista).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT
			f.id,
			f.nome,
			pgp_sym_decrypt(f.cpf, $1)::text AS cpf,
			COALESCE(f.telefone, ''),
			COALESCE(f.email, ''),
			COALESCE(f.cargo, ''),
			COALESCE(f.setor, ''),
			f.status::text,
			CASE WHEN m.id IS NOT NULL THEN 'motorista' ELSE 'funcionario' END AS tipo,
			(m.id IS NOT NULL) AS is_motorista,
			f.data_admissao,
			COALESCE(f.salario_base, 0),
			f.created_at
		FROM funcionarios f
		LEFT JOIN motoristas m ON m.id = f.id
		WHERE ($2 = '' OR f.nome ILIKE '%' || $2 || '%' OR COALESCE(f.email, '') ILIKE '%' || $2 || '%' OR COALESCE(f.cargo, '') ILIKE '%' || $2 || '%' OR COALESCE(f.setor, '') ILIKE '%' || $2 || '%')
		  AND ($3 = '' OR f.status::text = $3)
		  AND ($4 = '' OR CASE WHEN m.id IS NOT NULL THEN 'motorista' ELSE 'funcionario' END = $4)
		  AND ($5 OR m.id IS NULL)
		ORDER BY f.nome ASC
		LIMIT $6 OFFSET $7
	`

	rows, err := r.db.Query(
		ctx,
		query,
		r.encryptionKey,
		filter.Search,
		filter.Status,
		normalizeFuncionarioTipo(filter.Tipo),
		filter.IncludeMotorista,
		filter.Limit,
		(filter.Page-1)*filter.Limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.FuncionarioListItem, 0)
	for rows.Next() {
		var item domain.FuncionarioListItem
		var cpf string
		var dataAdmissao *time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Nome,
			&cpf,
			&item.Telefone,
			&item.Email,
			&item.Cargo,
			&item.Setor,
			&item.Status,
			&item.Tipo,
			&item.IsMotorista,
			&dataAdmissao,
			&item.SalarioBase,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		item.CPF = maskCPF(cpf)
		item.DataAdmissao = formatOptionalDate(dataAdmissao)
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *FuncionarioRepository) GetByID(ctx context.Context, id string) (*domain.FuncionarioDetail, error) {
	const query = `
		SELECT
			f.id,
			f.nome,
			pgp_sym_decrypt(f.cpf, $2)::text AS cpf,
			COALESCE(pgp_sym_decrypt(f.rg, $2)::text, ''),
			f.data_nascimento,
			COALESCE(f.telefone, ''),
			COALESCE(f.email, ''),
			COALESCE(f.cep, ''),
			COALESCE(f.endereco, ''),
			COALESCE(f.complemento, ''),
			COALESCE(f.numero, ''),
			COALESCE(f.bairro, ''),
			COALESCE(f.cidade, ''),
			COALESCE(f.estado, ''),
			COALESCE(f.cargo, ''),
			COALESCE(f.setor, ''),
			COALESCE(f.tipo_contrato::text, ''),
			f.data_admissao,
			f.data_demissao,
			f.status::text,
			COALESCE(f.salario_base, 0),
			COALESCE(f.tipo_pagamento::text, ''),
			COALESCE(f.valor_hora_extra, 0),
			COALESCE(f.adicional_noturno, 0),
			COALESCE(f.vale_alimentacao, 0),
			COALESCE(f.outros_descontos, 0),
			COALESCE(f.banco, ''),
			COALESCE(f.agencia, ''),
			COALESCE(f.conta, ''),
			COALESCE(f.tipo_conta::text, ''),
			COALESCE(f.chave_pix, ''),
			fcp.horario_entrada,
			fcp.horario_saida,
			fcp.horario_almoco,
			COALESCE(fcp.horas_extras, 0),
			COALESCE(fcp.faltas, 0),
			COALESCE(fcp.atestados, 0),
			COALESCE(f.observacoes, ''),
			CASE WHEN m.id IS NOT NULL THEN 'motorista' ELSE 'funcionario' END AS tipo,
			(m.id IS NOT NULL) AS is_motorista,
			f.created_at,
			f.updated_at
		FROM funcionarios f
		LEFT JOIN funcionario_controle_ponto fcp ON fcp.funcionario_id = f.id
		LEFT JOIN motoristas m ON m.id = f.id
		WHERE f.id = $1
		LIMIT 1
	`

	var item domain.FuncionarioDetail
	var dataNascimento *time.Time
	var dataAdmissao *time.Time
	var dataDemissao *time.Time
	var horarioEntrada *time.Time
	var horarioSaida *time.Time
	var horarioAlmoco *time.Time

	err := r.db.QueryRow(ctx, query, id, r.encryptionKey).Scan(
		&item.ID,
		&item.Nome,
		&item.CPF,
		&item.RG,
		&dataNascimento,
		&item.Telefone,
		&item.Email,
		&item.CEP,
		&item.Endereco,
		&item.Complemento,
		&item.Numero,
		&item.Bairro,
		&item.Cidade,
		&item.Estado,
		&item.Cargo,
		&item.Setor,
		&item.TipoContrato,
		&dataAdmissao,
		&dataDemissao,
		&item.Status,
		&item.SalarioBase,
		&item.TipoPagamento,
		&item.ValorHoraExtra,
		&item.AdicionalNoturno,
		&item.ValeAlimentacao,
		&item.OutrosDescontos,
		&item.Banco,
		&item.Agencia,
		&item.Conta,
		&item.TipoConta,
		&item.ChavePix,
		&horarioEntrada,
		&horarioSaida,
		&horarioAlmoco,
		&item.HorasExtras,
		&item.Faltas,
		&item.Atestados,
		&item.Observacoes,
		&item.Tipo,
		&item.IsMotorista,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	item.DataNascimento = formatOptionalDate(dataNascimento)
	item.DataAdmissao = formatOptionalDate(dataAdmissao)
	item.DataDemissao = formatOptionalDate(dataDemissao)
	item.HorarioEntrada = formatOptionalTime(horarioEntrada)
	item.HorarioSaida = formatOptionalTime(horarioSaida)
	item.HorarioAlmoco = formatOptionalTime(horarioAlmoco)

	return &item, nil
}

func (r *FuncionarioRepository) Create(ctx context.Context, input domain.FuncionarioCreateRequest) (*domain.FuncionarioDetail, error) {
	return r.createBase(ctx, nil, input)
}

func (r *FuncionarioRepository) Update(ctx context.Context, id string, input domain.FuncionarioUpdateRequest) (*domain.FuncionarioDetail, error) {
	return r.updateBase(ctx, nil, id, input)
}

func (r *FuncionarioRepository) Delete(ctx context.Context, id string) error {
	const hasMotoristaQuery = `SELECT EXISTS(SELECT 1 FROM motoristas WHERE id = $1)`

	var hasMotorista bool
	if err := r.db.QueryRow(ctx, hasMotoristaQuery, id).Scan(&hasMotorista); err != nil {
		return err
	}
	if hasMotorista {
		return domain.ErrProtectedRecord
	}

	const deleteQuery = `DELETE FROM funcionarios WHERE id = $1`
	tag, err := r.db.Exec(ctx, deleteQuery, id)
	if err != nil {
		return mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *FuncionarioRepository) UpdateStatus(ctx context.Context, id, status string) (*domain.FuncionarioDetail, error) {
	const query = `
		UPDATE funcionarios
		SET status = $2::status_funcionario
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id, normalizeFuncionarioStatus(status))
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *FuncionarioRepository) updateBase(ctx context.Context, tx pgx.Tx, id string, input domain.FuncionarioUpdateRequest) (*domain.FuncionarioDetail, error) {
	dataNascimento, err := parseOptionalDate(input.DataNascimento)
	if err != nil {
		return nil, err
	}

	dataAdmissao, err := parseOptionalDate(input.DataAdmissao)
	if err != nil {
		return nil, err
	}

	dataDemissao, err := parseOptionalDate(input.DataDemissao)
	if err != nil {
		return nil, err
	}

	horarioEntrada, err := parseOptionalTime(input.HorarioEntrada)
	if err != nil {
		return nil, err
	}

	horarioSaida, err := parseOptionalTime(input.HorarioSaida)
	if err != nil {
		return nil, err
	}

	horarioAlmoco, err := parseOptionalTime(input.HorarioAlmoco)
	if err != nil {
		return nil, err
	}

	cpf := normalizeDigits(input.CPF)
	if len(cpf) != 11 {
		return nil, domain.ErrInvalidInput
	}

	ownsTx := false
	if tx == nil {
		tx, err = r.db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return nil, err
		}
		ownsTx = true
		defer tx.Rollback(ctx)
	}

	const updateFuncionarioQuery = `
		UPDATE funcionarios
		SET
			nome = $2,
			cpf = pgp_sym_encrypt($3, $30),
			cpf_hash = encode(digest($3, 'sha256'), 'hex'),
			rg = CASE WHEN NULLIF($4, '') IS NULL THEN NULL ELSE pgp_sym_encrypt($4, $30) END,
			data_nascimento = $5,
			telefone = NULLIF($6, ''),
			email = NULLIF($7, ''),
			cep = NULLIF($8, ''),
			endereco = NULLIF($9, ''),
			complemento = NULLIF($10, ''),
			numero = NULLIF($11, ''),
			bairro = NULLIF($12, ''),
			cidade = NULLIF($13, ''),
			estado = NULLIF($14, ''),
			cargo = NULLIF($15, ''),
			setor = NULLIF($16, ''),
			tipo_contrato = $17::tipo_contrato_funcionario,
			data_admissao = $18,
			data_demissao = $19,
			status = $20::status_funcionario,
			salario_base = $21,
			tipo_pagamento = $22::tipo_pagamento_funcionario,
			valor_hora_extra = $23,
			adicional_noturno = $24,
			vale_alimentacao = $25,
			outros_descontos = $26,
			banco = NULLIF($27, ''),
			agencia = NULLIF($28, ''),
			conta = NULLIF($29, ''),
			tipo_conta = $31::tipo_conta_bancaria_funcionario,
			chave_pix = NULLIF($32, ''),
			observacoes = NULLIF($33, '')
		WHERE id = $1
	`

	tag, err := tx.Exec(
		ctx,
		updateFuncionarioQuery,
		id,
		strings.TrimSpace(input.Nome),
		cpf,
		normalizeDigits(input.RG),
		dataNascimento,
		strings.TrimSpace(input.Telefone),
		normalizeNullableEmail(input.Email),
		normalizeDigits(input.CEP),
		strings.TrimSpace(input.Endereco),
		strings.TrimSpace(input.Complemento),
		strings.TrimSpace(input.Numero),
		strings.TrimSpace(input.Bairro),
		strings.TrimSpace(input.Cidade),
		strings.ToUpper(strings.TrimSpace(input.Estado)),
		strings.TrimSpace(input.Cargo),
		strings.TrimSpace(input.Setor),
		normalizeTipoContrato(input.TipoContrato),
		dataAdmissao,
		dataDemissao,
		normalizeFuncionarioStatus(input.Status),
		input.SalarioBase,
		normalizeTipoPagamento(input.TipoPagamento),
		input.ValorHoraExtra,
		input.AdicionalNoturno,
		input.ValeAlimentacao,
		input.OutrosDescontos,
		strings.TrimSpace(input.Banco),
		strings.TrimSpace(input.Agencia),
		r.encryptionKey,
		strings.TrimSpace(input.Conta),
		normalizeTipoConta(input.TipoConta),
		strings.TrimSpace(input.ChavePix),
		strings.TrimSpace(input.Observacoes),
	)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}

	if err := r.upsertControlePonto(ctx, tx, id, horarioEntrada, horarioSaida, horarioAlmoco, input.HorasExtras, input.Faltas, input.Atestados); err != nil {
		return nil, err
	}

	if ownsTx {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return r.GetByID(ctx, id)
	}

	return &domain.FuncionarioDetail{ID: id}, nil
}

func (r *FuncionarioRepository) createBase(ctx context.Context, tx pgx.Tx, input domain.FuncionarioCreateRequest) (*domain.FuncionarioDetail, error) {
	dataNascimento, err := parseOptionalDate(input.DataNascimento)
	if err != nil {
		return nil, err
	}

	dataAdmissao, err := parseOptionalDate(input.DataAdmissao)
	if err != nil {
		return nil, err
	}

	dataDemissao, err := parseOptionalDate(input.DataDemissao)
	if err != nil {
		return nil, err
	}

	horarioEntrada, err := parseOptionalTime(input.HorarioEntrada)
	if err != nil {
		return nil, err
	}

	horarioSaida, err := parseOptionalTime(input.HorarioSaida)
	if err != nil {
		return nil, err
	}

	horarioAlmoco, err := parseOptionalTime(input.HorarioAlmoco)
	if err != nil {
		return nil, err
	}

	cpf := normalizeDigits(input.CPF)
	if len(cpf) != 11 {
		return nil, domain.ErrInvalidInput
	}

	ownsTx := false
	if tx == nil {
		tx, err = r.db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return nil, err
		}
		ownsTx = true
		defer tx.Rollback(ctx)
	}

	const insertFuncionarioQuery = `
		INSERT INTO funcionarios (
			nome, cpf, cpf_hash, rg, data_nascimento, telefone, email, cep, endereco, complemento, numero, bairro, cidade, estado,
			cargo, setor, tipo_contrato, data_admissao, data_demissao, status, salario_base, tipo_pagamento,
			valor_hora_extra, adicional_noturno, vale_alimentacao, outros_descontos,
			banco, agencia, conta, tipo_conta, chave_pix, observacoes
		)
		VALUES (
			$1, pgp_sym_encrypt($2, $30), encode(digest($2, 'sha256'), 'hex'),
			CASE WHEN NULLIF($3, '') IS NULL THEN NULL ELSE pgp_sym_encrypt($3, $30) END,
			$4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''), NULLIF($13, ''),
			NULLIF($14, ''), NULLIF($15, ''), $16::tipo_contrato_funcionario, $17, $18, $19::status_funcionario, $20, $21::tipo_pagamento_funcionario,
			$22, $23, $24, $25, NULLIF($26, ''), NULLIF($27, ''), NULLIF($28, ''), $31::tipo_conta_bancaria_funcionario, NULLIF($32, ''), NULLIF($33, '')
		)
		RETURNING id
	`

	var id string
	err = tx.QueryRow(
		ctx,
		insertFuncionarioQuery,
		strings.TrimSpace(input.Nome),
		cpf,
		normalizeDigits(input.RG),
		dataNascimento,
		strings.TrimSpace(input.Telefone),
		normalizeNullableEmail(input.Email),
		normalizeDigits(input.CEP),
		strings.TrimSpace(input.Endereco),
		strings.TrimSpace(input.Complemento),
		strings.TrimSpace(input.Numero),
		strings.TrimSpace(input.Bairro),
		strings.TrimSpace(input.Cidade),
		strings.ToUpper(strings.TrimSpace(input.Estado)),
		strings.TrimSpace(input.Cargo),
		strings.TrimSpace(input.Setor),
		normalizeTipoContrato(input.TipoContrato),
		dataAdmissao,
		dataDemissao,
		normalizeFuncionarioStatus(input.Status),
		input.SalarioBase,
		normalizeTipoPagamento(input.TipoPagamento),
		input.ValorHoraExtra,
		input.AdicionalNoturno,
		input.ValeAlimentacao,
		input.OutrosDescontos,
		strings.TrimSpace(input.Banco),
		strings.TrimSpace(input.Agencia),
		strings.TrimSpace(input.Conta),
		r.encryptionKey,
		normalizeTipoConta(input.TipoConta),
		strings.TrimSpace(input.ChavePix),
		strings.TrimSpace(input.Observacoes),
	).Scan(&id)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	if err := r.upsertControlePonto(ctx, tx, id, horarioEntrada, horarioSaida, horarioAlmoco, input.HorasExtras, input.Faltas, input.Atestados); err != nil {
		return nil, err
	}

	if ownsTx {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return r.GetByID(ctx, id)
	}

	return &domain.FuncionarioDetail{ID: id}, nil
}

func (r *FuncionarioRepository) upsertControlePonto(ctx context.Context, tx pgx.Tx, funcionarioID string, horarioEntrada, horarioSaida, horarioAlmoco *time.Time, horasExtras float64, faltas, atestados int) error {
	const query = `
		INSERT INTO funcionario_controle_ponto (
			funcionario_id, horario_entrada, horario_saida, horario_almoco, horas_extras, faltas, atestados
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (funcionario_id) DO UPDATE SET
			horario_entrada = EXCLUDED.horario_entrada,
			horario_saida = EXCLUDED.horario_saida,
			horario_almoco = EXCLUDED.horario_almoco,
			horas_extras = EXCLUDED.horas_extras,
			faltas = EXCLUDED.faltas,
			atestados = EXCLUDED.atestados,
			updated_at = NOW()
	`

	_, err := tx.Exec(ctx, query, funcionarioID, horarioEntrada, horarioSaida, horarioAlmoco, horasExtras, faltas, atestados)
	if err != nil {
		return mapDatabaseError(err)
	}

	return nil
}

func normalizeFuncionarioStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return "ativo"
	}
	return status
}

func normalizeFuncionarioTipo(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "todos" {
		return ""
	}
	return value
}

func normalizeTipoContrato(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "clt"
	}
	return value
}

func normalizeTipoPagamento(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "mensal"
	}
	return value
}

func normalizeTipoConta(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "corrente"
	}
	return value
}
