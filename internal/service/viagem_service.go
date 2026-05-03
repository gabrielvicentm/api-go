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

	if err := s.repo.EnsureMotoristaAtivo(ctx, input.MotoristaID); err != nil {
		return nil, err
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
			Descricao:     "Campo atualizado",
		}); err != nil {
			return nil, err
		}
	}

	return updated, nil
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
