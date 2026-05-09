package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (r *FuncionarioRepository) ListFolhaPagamento(ctx context.Context, filter domain.FolhaPagamentoListFilter) ([]domain.FolhaPagamentoResumo, error) {
	competencia, err := parseCompetencia(filter.Competencia)
	if err != nil {
		return nil, err
	}

	const query = `
		SELECT
			f.id,
			f.nome,
			COALESCE(f.cargo, ''),
			COALESCE(f.setor, ''),
			f.status::text,
			(fp.id IS NOT NULL) AS registro_existente,
			COALESCE(fp.status::text, 'aberta'),
			COALESCE(fp.salario_base_snapshot, COALESCE(f.salario_base, 0)),
			COALESCE(fp.valor_hora_extra_snapshot, COALESCE(f.valor_hora_extra, 0)),
			COALESCE(fp.vale_alimentacao_snapshot, COALESCE(f.vale_alimentacao, 0)),
			COALESCE(fp.outros_descontos_snapshot, COALESCE(f.outros_descontos, 0)),
			COALESCE(fp.dias_faltas, 0),
			COALESCE(fp.dias_atestado, 0),
			COALESCE(fp.dias_ferias, 0),
			COALESCE(fp.dias_afastamento, 0),
			COALESCE(fp.horas_extras_50, 0),
			COALESCE(fp.horas_extras_100, 0),
			COALESCE(fp.horas_adicional_noturno, 0),
			COALESCE(fp.bonus, 0),
			COALESCE(fp.comissoes, 0),
			COALESCE(fp.outros_proventos, 0),
			COALESCE(fp.adiantamentos, 0),
			COALESCE(fp.desconto_inss, 0),
			COALESCE(fp.desconto_irrf, 0),
			COALESCE(fp.desconto_vale_transporte, 0),
			COALESCE(fp.descontos_manuais, 0)
		FROM funcionarios f
		LEFT JOIN funcionario_folha_mensal fp
		  ON fp.funcionario_id = f.id
		 AND fp.competencia = @competencia
		WHERE (@search = '' OR f.nome ILIKE '%' || @search || '%' OR COALESCE(f.cargo, '') ILIKE '%' || @search || '%' OR COALESCE(f.setor, '') ILIKE '%' || @search || '%')
		  AND (@status = '' OR f.status::text = @status)
		ORDER BY f.nome ASC
	`

	rows, err := r.db.Query(ctx, query, pgx.NamedArgs{
		"competencia": competencia,
		"search":      strings.TrimSpace(filter.Search),
		"status":      strings.TrimSpace(filter.Status),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.FolhaPagamentoResumo, 0)
	competenciaLabel := formatCompetencia(competencia)

	for rows.Next() {
		var item domain.FolhaPagamentoResumo
		folha := domain.FolhaPagamentoCompetencia{Competencia: competenciaLabel}

		if err := rows.Scan(
			&item.FuncionarioID,
			&item.Nome,
			&item.Cargo,
			&item.Setor,
			&item.StatusFuncionario,
			&item.RegistroExistente,
			&item.StatusFolha,
			&folha.SalarioBaseSnapshot,
			&folha.ValorHoraExtraSnapshot,
			&folha.ValeAlimentacaoSnapshot,
			&folha.OutrosDescontosSnapshot,
			&folha.DiasFaltas,
			&folha.DiasAtestado,
			&folha.DiasFerias,
			&folha.DiasAfastamento,
			&folha.HorasExtras50,
			&folha.HorasExtras100,
			&folha.HorasAdicionalNoturno,
			&folha.Bonus,
			&folha.Comissoes,
			&folha.OutrosProventos,
			&folha.Adiantamentos,
			&folha.DescontoINSS,
			&folha.DescontoIRRF,
			&folha.DescontoValeTransporte,
			&folha.DescontosManuais,
		); err != nil {
			return nil, err
		}

		folha.Status = item.StatusFolha
		calculo, err := domain.CalculateFolhaPagamento(folha)
		if err != nil {
			return nil, err
		}

		item.Competencia = competenciaLabel
		item.SalarioBase = calculo.SalarioBase
		item.TotalProventos = calculo.TotalProventos
		item.TotalDescontos = calculo.TotalDescontos
		item.SalarioLiquido = calculo.SalarioLiquido
		item.DiasFaltas = folha.DiasFaltas
		item.DiasFerias = folha.DiasFerias
		item.HorasExtras50 = folha.HorasExtras50
		item.HorasExtras100 = folha.HorasExtras100
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *FuncionarioRepository) GetFolhaPagamento(ctx context.Context, id, competenciaRaw string) (*domain.FolhaPagamentoDetalhe, error) {
	competencia, err := parseCompetencia(competenciaRaw)
	if err != nil {
		return nil, err
	}

	const query = `
		SELECT
			f.id,
			f.nome,
			COALESCE(f.cargo, ''),
			COALESCE(f.setor, ''),
			f.status::text,
			COALESCE(f.tipo_pagamento::text, 'mensal'),
			(m.id IS NOT NULL) AS is_motorista,
			(fp.id IS NOT NULL) AS registro_existente,
			COALESCE(fp.salario_base_snapshot, COALESCE(f.salario_base, 0)),
			COALESCE(fp.valor_hora_extra_snapshot, COALESCE(f.valor_hora_extra, 0)),
			COALESCE(fp.vale_alimentacao_snapshot, COALESCE(f.vale_alimentacao, 0)),
			COALESCE(fp.outros_descontos_snapshot, COALESCE(f.outros_descontos, 0)),
			COALESCE(fp.dias_faltas, 0),
			COALESCE(fp.dias_atestado, 0),
			COALESCE(fp.dias_ferias, 0),
			COALESCE(fp.dias_afastamento, 0),
			COALESCE(fp.horas_extras_50, 0),
			COALESCE(fp.horas_extras_100, 0),
			COALESCE(fp.horas_adicional_noturno, 0),
			COALESCE(fp.bonus, 0),
			COALESCE(fp.comissoes, 0),
			COALESCE(fp.outros_proventos, 0),
			COALESCE(fp.adiantamentos, 0),
			COALESCE(fp.desconto_inss, 0),
			COALESCE(fp.desconto_irrf, 0),
			COALESCE(fp.desconto_vale_transporte, 0),
			COALESCE(fp.descontos_manuais, 0),
			COALESCE(fp.observacoes, ''),
			COALESCE(fp.status::text, 'aberta')
		FROM funcionarios f
		LEFT JOIN motoristas m ON m.id = f.id
		LEFT JOIN funcionario_folha_mensal fp
		  ON fp.funcionario_id = f.id
		 AND fp.competencia = @competencia
		WHERE f.id = @id
		LIMIT 1
	`

	item := &domain.FolhaPagamentoDetalhe{
		Folha: domain.FolhaPagamentoCompetencia{
			Competencia: formatCompetencia(competencia),
		},
	}

	err = r.db.QueryRow(ctx, query, pgx.NamedArgs{
		"id":          strings.TrimSpace(id),
		"competencia": competencia,
	}).Scan(
		&item.Funcionario.ID,
		&item.Funcionario.Nome,
		&item.Funcionario.Cargo,
		&item.Funcionario.Setor,
		&item.Funcionario.Status,
		&item.Funcionario.TipoPagamento,
		&item.Funcionario.IsMotorista,
		&item.RegistroExistente,
		&item.Folha.SalarioBaseSnapshot,
		&item.Folha.ValorHoraExtraSnapshot,
		&item.Folha.ValeAlimentacaoSnapshot,
		&item.Folha.OutrosDescontosSnapshot,
		&item.Folha.DiasFaltas,
		&item.Folha.DiasAtestado,
		&item.Folha.DiasFerias,
		&item.Folha.DiasAfastamento,
		&item.Folha.HorasExtras50,
		&item.Folha.HorasExtras100,
		&item.Folha.HorasAdicionalNoturno,
		&item.Folha.Bonus,
		&item.Folha.Comissoes,
		&item.Folha.OutrosProventos,
		&item.Folha.Adiantamentos,
		&item.Folha.DescontoINSS,
		&item.Folha.DescontoIRRF,
		&item.Folha.DescontoValeTransporte,
		&item.Folha.DescontosManuais,
		&item.Folha.Observacoes,
		&item.Folha.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	calculo, err := domain.CalculateFolhaPagamento(item.Folha)
	if err != nil {
		return nil, err
	}
	item.Calculo = calculo

	return item, nil
}

func (r *FuncionarioRepository) UpsertFolhaPagamento(ctx context.Context, id string, input domain.FolhaPagamentoUpsertRequest) (*domain.FolhaPagamentoDetalhe, error) {
	competencia, err := parseCompetencia(input.Competencia)
	if err != nil {
		return nil, err
	}

	if _, err := r.GetFolhaPagamento(ctx, id, input.Competencia); err != nil {
		return nil, err
	}

	folha := domain.FolhaPagamentoCompetencia{
		Competencia:             formatCompetencia(competencia),
		SalarioBaseSnapshot:     input.SalarioBaseSnapshot,
		ValorHoraExtraSnapshot:  input.ValorHoraExtraSnapshot,
		ValeAlimentacaoSnapshot: input.ValeAlimentacaoSnapshot,
		OutrosDescontosSnapshot: input.OutrosDescontosSnapshot,
		DiasFaltas:              input.DiasFaltas,
		DiasAtestado:            input.DiasAtestado,
		DiasFerias:              input.DiasFerias,
		DiasAfastamento:         input.DiasAfastamento,
		HorasExtras50:           input.HorasExtras50,
		HorasExtras100:          input.HorasExtras100,
		HorasAdicionalNoturno:   input.HorasAdicionalNoturno,
		Bonus:                   input.Bonus,
		Comissoes:               input.Comissoes,
		OutrosProventos:         input.OutrosProventos,
		Adiantamentos:           input.Adiantamentos,
		DescontoINSS:            input.DescontoINSS,
		DescontoIRRF:            input.DescontoIRRF,
		DescontoValeTransporte:  input.DescontoValeTransporte,
		DescontosManuais:        input.DescontosManuais,
		Observacoes:             strings.TrimSpace(input.Observacoes),
		Status:                  normalizeStatusFolha(input.Status),
	}

	if _, err := domain.CalculateFolhaPagamento(folha); err != nil {
		return nil, err
	}

	var pagoEm any
	if folha.Status == "paga" {
		pagoEm = time.Now()
	}

	const query = `
		INSERT INTO funcionario_folha_mensal (
			funcionario_id,
			competencia,
			salario_base_snapshot,
			valor_hora_extra_snapshot,
			vale_alimentacao_snapshot,
			outros_descontos_snapshot,
			dias_faltas,
			dias_atestado,
			dias_ferias,
			dias_afastamento,
			horas_extras_50,
			horas_extras_100,
			horas_adicional_noturno,
			bonus,
			comissoes,
			outros_proventos,
			adiantamentos,
			desconto_inss,
			desconto_irrf,
			desconto_vale_transporte,
			descontos_manuais,
			observacoes,
			status,
			pago_em
		)
		VALUES (
			@funcionario_id,
			@competencia,
			@salario_base_snapshot,
			@valor_hora_extra_snapshot,
			@vale_alimentacao_snapshot,
			@outros_descontos_snapshot,
			@dias_faltas,
			@dias_atestado,
			@dias_ferias,
			@dias_afastamento,
			@horas_extras_50,
			@horas_extras_100,
			@horas_adicional_noturno,
			@bonus,
			@comissoes,
			@outros_proventos,
			@adiantamentos,
			@desconto_inss,
			@desconto_irrf,
			@desconto_vale_transporte,
			@descontos_manuais,
			NULLIF(@observacoes, ''),
			@status::status_folha_funcionario,
			@pago_em
		)
		ON CONFLICT (funcionario_id, competencia) DO UPDATE SET
			salario_base_snapshot = EXCLUDED.salario_base_snapshot,
			valor_hora_extra_snapshot = EXCLUDED.valor_hora_extra_snapshot,
			vale_alimentacao_snapshot = EXCLUDED.vale_alimentacao_snapshot,
			outros_descontos_snapshot = EXCLUDED.outros_descontos_snapshot,
			dias_faltas = EXCLUDED.dias_faltas,
			dias_atestado = EXCLUDED.dias_atestado,
			dias_ferias = EXCLUDED.dias_ferias,
			dias_afastamento = EXCLUDED.dias_afastamento,
			horas_extras_50 = EXCLUDED.horas_extras_50,
			horas_extras_100 = EXCLUDED.horas_extras_100,
			horas_adicional_noturno = EXCLUDED.horas_adicional_noturno,
			bonus = EXCLUDED.bonus,
			comissoes = EXCLUDED.comissoes,
			outros_proventos = EXCLUDED.outros_proventos,
			adiantamentos = EXCLUDED.adiantamentos,
			desconto_inss = EXCLUDED.desconto_inss,
			desconto_irrf = EXCLUDED.desconto_irrf,
			desconto_vale_transporte = EXCLUDED.desconto_vale_transporte,
			descontos_manuais = EXCLUDED.descontos_manuais,
			observacoes = EXCLUDED.observacoes,
			status = EXCLUDED.status,
			pago_em = EXCLUDED.pago_em
	`

	if _, err := r.db.Exec(ctx, query, pgx.NamedArgs{
		"funcionario_id":            strings.TrimSpace(id),
		"competencia":               competencia,
		"salario_base_snapshot":     folha.SalarioBaseSnapshot,
		"valor_hora_extra_snapshot": folha.ValorHoraExtraSnapshot,
		"vale_alimentacao_snapshot": folha.ValeAlimentacaoSnapshot,
		"outros_descontos_snapshot": folha.OutrosDescontosSnapshot,
		"dias_faltas":               folha.DiasFaltas,
		"dias_atestado":             folha.DiasAtestado,
		"dias_ferias":               folha.DiasFerias,
		"dias_afastamento":          folha.DiasAfastamento,
		"horas_extras_50":           folha.HorasExtras50,
		"horas_extras_100":          folha.HorasExtras100,
		"horas_adicional_noturno":   folha.HorasAdicionalNoturno,
		"bonus":                     folha.Bonus,
		"comissoes":                 folha.Comissoes,
		"outros_proventos":          folha.OutrosProventos,
		"adiantamentos":             folha.Adiantamentos,
		"desconto_inss":             folha.DescontoINSS,
		"desconto_irrf":             folha.DescontoIRRF,
		"desconto_vale_transporte":  folha.DescontoValeTransporte,
		"descontos_manuais":         folha.DescontosManuais,
		"observacoes":               folha.Observacoes,
		"status":                    folha.Status,
		"pago_em":                   pagoEm,
	}); err != nil {
		return nil, mapDatabaseError(err)
	}

	return r.GetFolhaPagamento(ctx, id, input.Competencia)
}

func parseCompetencia(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}

	parsed, err := time.Parse("2006-01", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("competencia invalida: %w", domain.ErrInvalidInput)
	}

	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func formatCompetencia(value time.Time) string {
	return value.Format("2006-01")
}

func normalizeStatusFolha(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "aberta"
	}
	return value
}
