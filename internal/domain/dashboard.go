package domain

import "time"

type DashboardSnapshot struct {
	Summary   DashboardSummary    `json:"summary"`
	Metrics   DashboardMetrics    `json:"metrics"`
	Alerts    DashboardAlerts     `json:"alerts"`
	Activities []DashboardTripRow `json:"activities"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type DashboardSummary struct {
	TotalViagens               int64 `json:"total_viagens"`
	ViagensHoje                int64 `json:"viagens_hoje"`
	ViagensEmAndamento         int64 `json:"viagens_em_andamento"`
	ViagensPendentes           int64 `json:"viagens_pendentes"`
	ViagensAtrasadas           int64 `json:"viagens_atrasadas"`
	VeiculosEmUso              int64 `json:"veiculos_em_uso"`
	VeiculosIndisponiveis      int64 `json:"veiculos_indisponiveis"`
	ManutencoesEmAndamento     int64 `json:"manutencoes_em_andamento"`
	MotoristasAtivos           int64 `json:"motoristas_ativos"`
	MotoristasCNHVencendo      int64 `json:"motoristas_cnh_vencendo"`
	AlertasPendenciasTotal     int64 `json:"alertas_pendencias_total"`
	AlertasCriticosTotal       int64 `json:"alertas_criticos_total"`
}

type DashboardMetrics struct {
	GastoOperacionalHoje   float64 `json:"gasto_operacional_hoje"`
	GastoAbastecimentoHoje float64 `json:"gasto_abastecimento_hoje"`
	GastoManutencaoHoje    float64 `json:"gasto_manutencao_hoje"`
	ViagensConcluidasHoje  int64   `json:"viagens_concluidas_hoje"`
	AbastecimentosHoje     int64   `json:"abastecimentos_hoje"`
	FinalizacoesPendentes  int64   `json:"finalizacoes_pendentes"`
	OcorrenciasHoje        int64   `json:"ocorrencias_hoje"`
	ParadasAbertas         int64   `json:"paradas_abertas"`
	DisponibilidadeFrota   float64 `json:"disponibilidade_frota"`
}

type DashboardAlerts struct {
	Operational []string `json:"operational"`
	Fleet       []string `json:"fleet"`
}

type DashboardTripRow struct {
	ID           string `json:"id"`
	Vehicle      string `json:"vehicle"`
	Driver       string `json:"driver"`
	Route        string `json:"route"`
	Status       string `json:"status"`
}
