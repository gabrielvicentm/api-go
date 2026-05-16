package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
)

type ViagemService struct {
	repo            *repository.ViagemRepository
	documentStorage ViagemDocumentStorage
}

func NewViagemService(repo *repository.ViagemRepository, documentStorage ViagemDocumentStorage) *ViagemService {
	return &ViagemService{
		repo:            repo,
		documentStorage: documentStorage,
	}
}

func (s *ViagemService) Create(ctx context.Context, input domain.ViagemCreateRequest, actorType, actorID string) (*domain.ViagemDetail, error) {
	if err := s.repo.EnsureMotoristaAtivo(ctx, input.MotoristaID); err != nil {
		return nil, err
	}
	if err := s.repo.EnsureVeiculoDisponivel(ctx, input.VeiculoID); err != nil {
		return nil, err
	}
	if err := s.repo.ValidateKMInicial(ctx, input.VeiculoID, input.KMInicial); err != nil {
		return nil, err
	}

	item, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateHistory(ctx, domain.ViagemHistoricoCreateInput{
		ViagemID:    item.ID,
		UsuarioTipo: actorType,
		UsuarioID:   actorID,
		Descricao:   "Viagem criada",
	}); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ViagemService) Update(ctx context.Context, id string, input domain.ViagemUpdateRequest, actorType, actorID string) (*domain.ViagemDetail, error) {
	before, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateViagemStatus(input.Status); err != nil {
		return nil, err
	}
	if err := s.repo.EnsureMotoristaAtivo(ctx, input.MotoristaID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.VeiculoID) != before.VeiculoID {
		if err := s.repo.EnsureVeiculoDisponivel(ctx, input.VeiculoID); err != nil {
			return nil, err
		}
	}
	if err := s.repo.ValidateKMInicial(ctx, input.VeiculoID, input.KMInicial); err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	for _, change := range collectViagemChanges(before, updated) {
		if err := s.repo.CreateHistory(ctx, domain.ViagemHistoricoCreateInput{
			ViagemID:      id,
			UsuarioTipo:   actorType,
			UsuarioID:     actorID,
			CampoAlterado: change.Field,
			ValorAnterior: change.Before,
			ValorNovo:     change.After,
			Descricao:     describeViagemChange(change),
		}); err != nil {
			return nil, err
		}
	}

	return updated, nil
}

func (s *ViagemService) FinalizeByAdmin(ctx context.Context, id string, input domain.ViagemFinalizacaoAdminRequest, actorType, actorID string) (*domain.ViagemDetail, error) {
	before, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updated, err := s.repo.FinalizeByAdmin(ctx, id, input)
	if err != nil {
		return nil, err
	}

	for _, change := range collectViagemChanges(before, updated) {
		if err := s.repo.CreateHistory(ctx, domain.ViagemHistoricoCreateInput{
			ViagemID:      id,
			UsuarioTipo:   actorType,
			UsuarioID:     actorID,
			CampoAlterado: change.Field,
			ValorAnterior: change.Before,
			ValorNovo:     change.After,
			Descricao:     describeViagemChange(change),
		}); err != nil {
			return nil, err
		}
	}

	if err := s.repo.CreateHistory(ctx, domain.ViagemHistoricoCreateInput{
		ViagemID:    id,
		UsuarioTipo: actorType,
		UsuarioID:   actorID,
		Descricao:   "Viagem finalizada pelo administrativo",
	}); err != nil {
		return nil, err
	}

	return updated, nil
}

func validateViagemStatus(status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return nil
	}

	switch status {
	case "pendente", "em_andamento", "parada", "concluida", "cancelada":
		return nil
	default:
		return fmt.Errorf("status de viagem invalido: %w", domain.ErrInvalidInput)
	}
}

func (s *ViagemService) StartStop(ctx context.Context, motoristaID string, input domain.ViagemParadaStartRequest, actorType, actorID string) (*domain.ViagemParadaStateResponse, error) {
	input.Motivo = strings.TrimSpace(input.Motivo)
	if input.Motivo == "" {
		return nil, fmt.Errorf("motivo da parada obrigatorio: %w", domain.ErrInvalidInput)
	}

	trip, err := s.currentMotoristaTrip(ctx, motoristaID)
	if err != nil {
		return nil, err
	}

	stop, updatedTrip, err := s.repo.StartStop(ctx, trip.ID, input)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateHistory(ctx, domain.ViagemHistoricoCreateInput{
		ViagemID:      updatedTrip.ID,
		UsuarioTipo:   actorType,
		UsuarioID:     actorID,
		CampoAlterado: "status",
		ValorAnterior: trip.Status,
		ValorNovo:     updatedTrip.Status,
		Descricao:     fmt.Sprintf("Motorista registrou parada. Motivo: %s", stop.Motivo),
	}); err != nil {
		return nil, err
	}

	return &domain.ViagemParadaStateResponse{
		Viagem: updatedTrip,
		Parada: stop,
	}, nil
}

func (s *ViagemService) FinishStop(ctx context.Context, motoristaID string, actorType, actorID string) (*domain.ViagemParadaStateResponse, error) {
	trip, err := s.currentMotoristaTrip(ctx, motoristaID)
	if err != nil {
		return nil, err
	}

	stop, updatedTrip, err := s.repo.FinishOpenStop(ctx, trip.ID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateHistory(ctx, domain.ViagemHistoricoCreateInput{
		ViagemID:      updatedTrip.ID,
		UsuarioTipo:   actorType,
		UsuarioID:     actorID,
		CampoAlterado: "status",
		ValorAnterior: trip.Status,
		ValorNovo:     updatedTrip.Status,
		Descricao:     "Motorista voltou para a estrada",
	}); err != nil {
		return nil, err
	}

	return &domain.ViagemParadaStateResponse{
		Viagem: updatedTrip,
		Parada: stop,
	}, nil
}

func (s *ViagemService) currentMotoristaTrip(ctx context.Context, motoristaID string) (*domain.ViagemDetail, error) {
	items, _, err := s.repo.List(ctx, domain.ViagemListFilter{
		MotoristaID:      motoristaID,
		ExcludeConcluded: true,
		Page:             1,
		Limit:            1,
	})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}

	return &items[0], nil
}

func (s *ViagemService) UploadDocument(ctx context.Context, viagemID string, body io.Reader, filename, documentType, contentType string, size int64, actorType, actorID string) (*domain.ViagemDocumentoItem, error) {
	if s.documentStorage == nil {
		return nil, fmt.Errorf("armazenamento de documentos nao configurado: %w", domain.ErrInvalidInput)
	}

	if _, err := s.repo.GetByID(ctx, viagemID); err != nil {
		return nil, err
	}

	documentURL, err := s.documentStorage.UploadViagemDocument(ctx, body, viagemID, filename, contentType)
	if err != nil {
		return nil, err
	}

	item, err := s.repo.CreateDocument(ctx, domain.ViagemDocumentoCreateInput{
		ViagemID:     viagemID,
		Nome:         filename,
		Tipo:         documentType,
		URL:          documentURL,
		TamanhoBytes: size,
	})
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateHistory(ctx, domain.ViagemHistoricoCreateInput{
		ViagemID:    viagemID,
		UsuarioTipo: actorType,
		UsuarioID:   actorID,
		Descricao:   "Documento de viagem enviado",
	}); err != nil {
		return nil, err
	}

	return item, nil
}

type viagemChange struct {
	Field  string
	Before string
	After  string
}

func describeViagemChange(change viagemChange) string {
	return fmt.Sprintf(
		"%s alterado de %q para %q",
		viagemFieldLabel(change.Field),
		formatHistoryValue(change.Before),
		formatHistoryValue(change.After),
	)
}

func viagemFieldLabel(field string) string {
	labels := map[string]string{
		"motorista_id":          "Motorista",
		"veiculo_id":            "Veiculo",
		"cliente_id":            "Cliente",
		"origem_cidade":         "Cidade de origem",
		"origem_uf":             "UF de origem",
		"destino_cidade":        "Cidade de destino",
		"destino_uf":            "UF de destino",
		"data_saida":            "Data de saida",
		"data_chegada_prevista": "Data de chegada prevista",
		"data_chegada_real":     "Data de chegada real",
		"distancia_km":          "Distancia em KM",
		"tipo_carga_id":         "Tipo de carga",
		"peso_carga_kg":         "Peso da carga",
		"valor_frete":           "Valor do frete",
		"km_inicial":            "KM inicial",
		"km_final":              "KM final",
		"status":                "Status",
		"observacoes":           "Observacoes",
	}

	if label, ok := labels[field]; ok {
		return label
	}

	return field
}

func formatHistoryValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "vazio"
	}

	return value
}

func collectViagemChanges(before, after *domain.ViagemDetail) []viagemChange {
	changes := make([]viagemChange, 0)
	appendViagemChange(&changes, "motorista_id", before.MotoristaID, after.MotoristaID)
	appendViagemChange(&changes, "veiculo_id", before.VeiculoID, after.VeiculoID)
	appendViagemChange(&changes, "cliente_id", before.ClienteID, after.ClienteID)
	appendViagemChange(&changes, "origem_cidade", before.OrigemCidade, after.OrigemCidade)
	appendViagemChange(&changes, "origem_uf", before.OrigemUF, after.OrigemUF)
	appendViagemChange(&changes, "destino_cidade", before.DestinoCidade, after.DestinoCidade)
	appendViagemChange(&changes, "destino_uf", before.DestinoUF, after.DestinoUF)
	appendViagemChange(&changes, "data_saida", formatOptionalTime(before.DataSaida), formatOptionalTime(after.DataSaida))
	appendViagemChange(&changes, "data_chegada_prevista", formatOptionalTime(before.DataChegadaPrevista), formatOptionalTime(after.DataChegadaPrevista))
	appendViagemChange(&changes, "data_chegada_real", formatOptionalTime(before.DataChegadaReal), formatOptionalTime(after.DataChegadaReal))
	appendViagemChange(&changes, "distancia_km", before.DistanciaKM, after.DistanciaKM)
	appendViagemChange(&changes, "tipo_carga_id", before.TipoCargaID, after.TipoCargaID)
	appendViagemChange(&changes, "peso_carga_kg", before.PesoCargaKG, after.PesoCargaKG)
	appendViagemChange(&changes, "valor_frete", before.ValorFrete, after.ValorFrete)
	appendViagemChange(&changes, "km_inicial", before.KMInicial, after.KMInicial)
	appendViagemChange(&changes, "km_final", before.KMFinal, after.KMFinal)
	appendViagemChange(&changes, "status", before.Status, after.Status)
	appendViagemChange(&changes, "observacoes", before.Observacoes, after.Observacoes)

	return changes
}

func appendViagemChange(changes *[]viagemChange, field, before, after string) {
	before = strings.TrimSpace(before)
	after = strings.TrimSpace(after)
	if before == after {
		return
	}

	*changes = append(*changes, viagemChange{
		Field:  field,
		Before: before,
		After:  after,
	})
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
