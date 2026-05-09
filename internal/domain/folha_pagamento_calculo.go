package domain

import (
	"fmt"
	"math"
	"strings"
)

const (
	payrollMonthlyDays  = 30.0
	payrollMonthlyHours = 220.0
)

func CalculateFolhaPagamento(folha FolhaPagamentoCompetencia) (FolhaPagamentoCalculo, error) {
	if err := validateFolhaPagamento(folha); err != nil {
		return FolhaPagamentoCalculo{}, err
	}

	salarioBase := folha.SalarioBaseSnapshot
	valorDia := salarioBase / payrollMonthlyDays
	valorHora := salarioBase / payrollMonthlyHours

	valorHoraExtra50Unit := folha.ValorHoraExtraSnapshot
	if valorHoraExtra50Unit <= 0 {
		valorHoraExtra50Unit = valorHora * 1.5
	}

	valorHoraExtra100Unit := valorHora * 2
	valorAdicionalNoturnoUnit := valorHora * 0.2

	descontoFaltas := valorDia * float64(folha.DiasFaltas)
	descontoAfastamento := valorDia * float64(folha.DiasAfastamento)
	descontoFerias := valorDia * float64(folha.DiasFerias)

	valorFerias := descontoFerias
	tercoFerias := valorFerias / 3
	valorHoraExtra50 := folha.HorasExtras50 * valorHoraExtra50Unit
	valorHoraExtra100 := folha.HorasExtras100 * valorHoraExtra100Unit
	valorAdicionalNoturno := folha.HorasAdicionalNoturno * valorAdicionalNoturnoUnit

	totalProventos := salarioBase +
		valorFerias +
		tercoFerias +
		valorHoraExtra50 +
		valorHoraExtra100 +
		valorAdicionalNoturno +
		folha.ValeAlimentacaoSnapshot +
		folha.Bonus +
		folha.Comissoes +
		folha.OutrosProventos

	totalDescontos := descontoFaltas +
		descontoAfastamento +
		descontoFerias +
		folha.OutrosDescontosSnapshot +
		folha.Adiantamentos +
		folha.DescontoINSS +
		folha.DescontoIRRF +
		folha.DescontoValeTransporte +
		folha.DescontosManuais

	return FolhaPagamentoCalculo{
		SalarioBase:           roundMoney(salarioBase),
		ValorDia:              roundMoney(valorDia),
		ValorHora:             roundMoney(valorHora),
		ValorHoraExtra50:      roundMoney(valorHoraExtra50),
		ValorHoraExtra100:     roundMoney(valorHoraExtra100),
		ValorAdicionalNoturno: roundMoney(valorAdicionalNoturno),
		ValorFerias:           roundMoney(valorFerias),
		TercoFerias:           roundMoney(tercoFerias),
		DescontoFaltas:        roundMoney(descontoFaltas),
		DescontoAfastamento:   roundMoney(descontoAfastamento),
		DescontoFerias:        roundMoney(descontoFerias),
		TotalProventos:        roundMoney(totalProventos),
		TotalDescontos:        roundMoney(totalDescontos),
		SalarioLiquido:        roundMoney(totalProventos - totalDescontos),
	}, nil
}

func validateFolhaPagamento(folha FolhaPagamentoCompetencia) error {
	if strings.TrimSpace(folha.Competencia) == "" {
		return fmt.Errorf("competencia obrigatoria: %w", ErrInvalidInput)
	}

	if folha.DiasFaltas < 0 || folha.DiasAtestado < 0 || folha.DiasFerias < 0 || folha.DiasAfastamento < 0 {
		return fmt.Errorf("dias informados nao podem ser negativos: %w", ErrInvalidInput)
	}

	if folha.HorasExtras50 < 0 || folha.HorasExtras100 < 0 || folha.HorasAdicionalNoturno < 0 {
		return fmt.Errorf("horas informadas nao podem ser negativas: %w", ErrInvalidInput)
	}

	if folha.SalarioBaseSnapshot < 0 || folha.ValorHoraExtraSnapshot < 0 || folha.ValeAlimentacaoSnapshot < 0 || folha.OutrosDescontosSnapshot < 0 {
		return fmt.Errorf("valores fixos da folha nao podem ser negativos: %w", ErrInvalidInput)
	}

	if folha.Bonus < 0 || folha.Comissoes < 0 || folha.OutrosProventos < 0 || folha.Adiantamentos < 0 || folha.DescontoINSS < 0 || folha.DescontoIRRF < 0 || folha.DescontoValeTransporte < 0 || folha.DescontosManuais < 0 {
		return fmt.Errorf("proventos e descontos nao podem ser negativos: %w", ErrInvalidInput)
	}

	if folha.DiasFaltas+folha.DiasAtestado+folha.DiasFerias+folha.DiasAfastamento > 31 {
		return fmt.Errorf("o total de dias informados excede o limite mensal: %w", ErrInvalidInput)
	}

	switch strings.TrimSpace(strings.ToLower(folha.Status)) {
	case "", "aberta", "fechada", "paga":
		return nil
	default:
		return fmt.Errorf("status da folha invalido: %w", ErrInvalidInput)
	}
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
