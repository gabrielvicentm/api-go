package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardRepository struct {
	db *pgxpool.Pool
}

func NewDashboardRepository(db *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetSnapshot(ctx context.Context) (*domain.DashboardSnapshot, error) {
	const summaryQuery = `
		SELECT
			(SELECT COUNT(*) FROM viagens) AS total_viagens,
			(SELECT total_viagens_hoje FROM vw_dashboard_hoje) AS viagens_hoje,
			(SELECT viagens_em_andamento FROM vw_dashboard_hoje) AS viagens_em_andamento,
			(SELECT viagens_pendentes FROM vw_dashboard_hoje) AS viagens_pendentes,
			(
				SELECT COUNT(*)
				FROM viagens
				WHERE status IN ('pendente', 'em_andamento', 'parada')
				  AND data_chegada_prevista IS NOT NULL
				  AND data_chegada_prevista < NOW()
			) AS viagens_atrasadas,
			(SELECT veiculos_em_uso FROM vw_dashboard_hoje) AS veiculos_em_uso,
			(
				(SELECT veiculos_em_manutencao FROM vw_dashboard_hoje)
				+ (SELECT COUNT(*) FROM veiculos WHERE status = 'inativo')
			) AS veiculos_indisponiveis,
			(SELECT COUNT(*) FROM manutencoes WHERE status = 'em_andamento') AS manutencoes_em_andamento,
			(SELECT COUNT(*) FROM motoristas m JOIN funcionarios f ON f.id = m.id WHERE f.status = 'ativo') AS motoristas_ativos,
			(SELECT COUNT(*) FROM vw_alertas WHERE tipo_alerta = 'cnh_vencimento') AS motoristas_cnh_vencendo,
			(SELECT COUNT(*) FROM vw_alertas) AS alertas_pendencias_total,
			(SELECT COUNT(*) FROM vw_alertas WHERE dias_restantes IS NOT NULL AND dias_restantes <= 7) AS alertas_criticos_total,
			(SELECT gasto_abastecimento_hoje FROM vw_dashboard_hoje) AS gasto_abastecimento_hoje,
			(SELECT gasto_manutencao_hoje FROM vw_dashboard_hoje) AS gasto_manutencao_hoje,
			(SELECT viagens_concluidas_hoje FROM vw_dashboard_hoje) AS viagens_concluidas_hoje,
			(SELECT COUNT(*) FROM abastecimentos WHERE DATE(registrado_em) = CURRENT_DATE) AS abastecimentos_hoje,
			(SELECT COUNT(*) FROM viagem_finalizacoes WHERE status = 'pendente') AS finalizacoes_pendentes,
			(SELECT COUNT(*) FROM ocorrencias WHERE DATE(registrado_em) = CURRENT_DATE) AS ocorrencias_hoje,
			(SELECT COUNT(*) FROM viagem_paradas WHERE finalizada_em IS NULL) AS paradas_abertas,
			(SELECT COUNT(*) FROM veiculos) AS total_veiculos,
			(SELECT veiculos_disponiveis FROM vw_dashboard_hoje) AS veiculos_disponiveis
	`

	var snapshot domain.DashboardSnapshot
	var totalVeiculos int64
	var veiculosDisponiveis int64

	err := r.db.QueryRow(ctx, summaryQuery).Scan(
		&snapshot.Summary.TotalViagens,
		&snapshot.Summary.ViagensHoje,
		&snapshot.Summary.ViagensEmAndamento,
		&snapshot.Summary.ViagensPendentes,
		&snapshot.Summary.ViagensAtrasadas,
		&snapshot.Summary.VeiculosEmUso,
		&snapshot.Summary.VeiculosIndisponiveis,
		&snapshot.Summary.ManutencoesEmAndamento,
		&snapshot.Summary.MotoristasAtivos,
		&snapshot.Summary.MotoristasCNHVencendo,
		&snapshot.Summary.AlertasPendenciasTotal,
		&snapshot.Summary.AlertasCriticosTotal,
		&snapshot.Metrics.GastoAbastecimentoHoje,
		&snapshot.Metrics.GastoManutencaoHoje,
		&snapshot.Metrics.ViagensConcluidasHoje,
		&snapshot.Metrics.AbastecimentosHoje,
		&snapshot.Metrics.FinalizacoesPendentes,
		&snapshot.Metrics.OcorrenciasHoje,
		&snapshot.Metrics.ParadasAbertas,
		&totalVeiculos,
		&veiculosDisponiveis,
	)
	if err != nil {
		return nil, err
	}

	snapshot.Metrics.GastoOperacionalHoje = snapshot.Metrics.GastoAbastecimentoHoje + snapshot.Metrics.GastoManutencaoHoje
	if totalVeiculos > 0 {
		snapshot.Metrics.DisponibilidadeFrota = (float64(veiculosDisponiveis) / float64(totalVeiculos)) * 100
	}

	operationalAlerts := buildOperationalAlerts(snapshot)
	fleetAlerts, err := r.listFleetAlerts(ctx)
	if err != nil {
		return nil, err
	}

	activities, err := r.listActivities(ctx)
	if err != nil {
		return nil, err
	}

	snapshot.Alerts = domain.DashboardAlerts{
		Operational: operationalAlerts,
		Fleet:       fleetAlerts,
	}
	snapshot.Activities = activities
	snapshot.UpdatedAt = time.Now().UTC()

	return &snapshot, nil
}

func (r *DashboardRepository) listFleetAlerts(ctx context.Context) ([]string, error) {
	const query = `
		SELECT tipo_alerta, descricao, data_referencia, dias_restantes
		FROM vw_alertas
		ORDER BY dias_restantes NULLS LAST, data_referencia NULLS LAST
		LIMIT 4
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0, 4)
	for rows.Next() {
		var tipoAlerta string
		var descricao string
		var dataReferencia *time.Time
		var diasRestantes *int

		if err := rows.Scan(&tipoAlerta, &descricao, &dataReferencia, &diasRestantes); err != nil {
			return nil, err
		}

		items = append(items, formatFleetAlert(tipoAlerta, descricao, dataReferencia, diasRestantes))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return []string{"Nenhum alerta de frota, documentacao ou manutencao no momento."}, nil
	}

	return items, nil
}

func (r *DashboardRepository) listActivities(ctx context.Context) ([]domain.DashboardTripRow, error) {
	const query = `
		SELECT
			v.id,
			ve.placa || ' - ' || ve.modelo AS vehicle,
			f.nome AS driver,
			v.origem_cidade || '/' || v.origem_uf || ' > ' || v.destino_cidade || '/' || v.destino_uf AS route,
			v.status::text
		FROM viagens v
		JOIN veiculos ve ON ve.id = v.veiculo_id
		JOIN motoristas m ON m.id = v.motorista_id
		JOIN funcionarios f ON f.id = m.id
		WHERE v.status IN ('pendente', 'em_andamento', 'parada')
		ORDER BY v.data_saida DESC NULLS LAST, v.created_at DESC
		LIMIT 8
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.DashboardTripRow, 0, 8)
	for rows.Next() {
		var item domain.DashboardTripRow
		if err := rows.Scan(&item.ID, &item.Vehicle, &item.Driver, &item.Route, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func buildOperationalAlerts(snapshot domain.DashboardSnapshot) []string {
	items := make([]string, 0, 5)

	if snapshot.Metrics.FinalizacoesPendentes > 0 {
		items = append(items, fmt.Sprintf("%d solicitacoes de finalizacao aguardando analise.", snapshot.Metrics.FinalizacoesPendentes))
	}
	if snapshot.Summary.ViagensAtrasadas > 0 {
		items = append(items, fmt.Sprintf("%d viagens estao atrasadas em relacao a previsao de chegada.", snapshot.Summary.ViagensAtrasadas))
	}
	if snapshot.Summary.ViagensPendentes > 0 {
		items = append(items, fmt.Sprintf("%d viagens pendentes de despacho.", snapshot.Summary.ViagensPendentes))
	}
	if snapshot.Summary.ViagensEmAndamento > 0 {
		items = append(items, fmt.Sprintf("%d viagens em andamento na operacao agora.", snapshot.Summary.ViagensEmAndamento))
	}
	if snapshot.Metrics.OcorrenciasHoje > 0 {
		items = append(items, fmt.Sprintf("%d ocorrencias registradas hoje.", snapshot.Metrics.OcorrenciasHoje))
	}
	if snapshot.Metrics.ParadasAbertas > 0 {
		items = append(items, fmt.Sprintf("%d paradas abertas exigem acompanhamento.", snapshot.Metrics.ParadasAbertas))
	}

	if len(items) == 0 {
		return []string{"Nenhuma pendencia operacional aberta no momento."}
	}

	return items
}

func formatFleetAlert(tipoAlerta, descricao string, dataReferencia *time.Time, diasRestantes *int) string {
	switch tipoAlerta {
	case "cnh_vencimento":
		return formatDeadlineAlert(descricao, "CNH vence", dataReferencia, diasRestantes)
	case "seguro_vencimento":
		return formatDeadlineAlert(descricao, "Seguro vence", dataReferencia, diasRestantes)
	case "licenciamento_vencimento":
		return formatDeadlineAlert(descricao, "Licenciamento vence", dataReferencia, diasRestantes)
	case "manutencao_preventiva":
		return fmt.Sprintf("%s esta proximo da manutencao preventiva por KM.", descricao)
	default:
		return descricao
	}
}

func formatDeadlineAlert(descricao, prefix string, dataReferencia *time.Time, diasRestantes *int) string {
	if dataReferencia == nil {
		return fmt.Sprintf("%s: %s.", prefix, descricao)
	}

	dateLabel := dataReferencia.Format("02/01/2006")
	if diasRestantes == nil {
		return fmt.Sprintf("%s: %s em %s.", prefix, descricao, dateLabel)
	}

	if *diasRestantes == 0 {
		return fmt.Sprintf("%s: %s hoje (%s).", prefix, descricao, dateLabel)
	}

	return fmt.Sprintf("%s: %s em %d dia(s) (%s).", prefix, descricao, *diasRestantes, dateLabel)
}
