package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
)

type OcorrenciaPhotoStorage interface {
	UploadOcorrenciaPhoto(ctx context.Context, body io.Reader, originalFilename, contentType string) (string, error)
}

type OcorrenciaService struct {
	repo         *repository.OcorrenciaRepository
	viagemRepo   *repository.ViagemRepository
	photoStorage OcorrenciaPhotoStorage
}

func NewOcorrenciaService(repo *repository.OcorrenciaRepository, viagemRepo *repository.ViagemRepository, photoStorage OcorrenciaPhotoStorage) *OcorrenciaService {
	return &OcorrenciaService{
		repo:         repo,
		viagemRepo:   viagemRepo,
		photoStorage: photoStorage,
	}
}

func (s *OcorrenciaService) Create(ctx context.Context, motoristaID string, input domain.OcorrenciaCreateRequest, photo io.Reader, photoFilename, photoContentType string) (*domain.OcorrenciaItem, error) {
	input.Tipo = normalizeOccurrenceType(input.Tipo)
	input.Motivo = strings.TrimSpace(input.Motivo)
	input.Descricao = strings.TrimSpace(input.Descricao)
	input.Latitude = strings.TrimSpace(input.Latitude)
	input.Longitude = strings.TrimSpace(input.Longitude)

	if input.Motivo == "" {
		return nil, fmt.Errorf("motivo da ocorrencia obrigatorio: %w", domain.ErrInvalidInput)
	}
	if input.Descricao == "" {
		return nil, fmt.Errorf("descricao da ocorrencia obrigatoria: %w", domain.ErrInvalidInput)
	}
	if err := validateMotoristaOccurrenceType(input.Tipo); err != nil {
		return nil, err
	}

	trip, err := s.currentMotoristaTrip(ctx, motoristaID)
	if err != nil {
		return nil, err
	}
	if trip.Status != "em_andamento" && trip.Status != "parada" {
		return nil, fmt.Errorf("viagem precisa estar em andamento ou parada para registrar ocorrencia: %w", domain.ErrInvalidInput)
	}

	photoURL := ""
	if photo != nil {
		if s.photoStorage == nil {
			return nil, fmt.Errorf("armazenamento de fotos nao configurado: %w", domain.ErrInvalidInput)
		}

		uploadedURL, err := s.photoStorage.UploadOcorrenciaPhoto(ctx, photo, photoFilename, photoContentType)
		if err != nil {
			return nil, err
		}
		photoURL = uploadedURL
	}

	item, err := s.repo.Create(ctx, domain.OcorrenciaCreateInput{
		ViagemID:    trip.ID,
		VeiculoID:   trip.VeiculoID,
		MotoristaID: motoristaID,
		Tipo:        input.Tipo,
		Motivo:      input.Motivo,
		Descricao:   input.Descricao,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		FotoURL:     photoURL,
	})
	if err != nil {
		return nil, err
	}

	if err := s.viagemRepo.CreateHistory(ctx, domain.ViagemHistoricoCreateInput{
		ViagemID:    trip.ID,
		UsuarioTipo: domain.ActorTypeMotorista,
		UsuarioID:   motoristaID,
		Descricao:   fmt.Sprintf("Motorista registrou ocorrencia. Motivo: %s", item.Motivo),
	}); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *OcorrenciaService) List(ctx context.Context, filter domain.OcorrenciaListFilter) ([]domain.OcorrenciaItem, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *OcorrenciaService) GetByID(ctx context.Context, id string) (*domain.OcorrenciaItem, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *OcorrenciaService) currentMotoristaTrip(ctx context.Context, motoristaID string) (*domain.ViagemDetail, error) {
	items, _, err := s.viagemRepo.List(ctx, domain.ViagemListFilter{
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

func normalizeOccurrenceType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "outro"
	}

	return value
}

func validateMotoristaOccurrenceType(value string) error {
	switch value {
	case "acidente", "pane_mecanica", "pane_eletrica", "furto", "avaria_carga", "outro":
		return nil
	case "multa", "atraso":
		return fmt.Errorf("tipo de ocorrencia nao permitido para o motorista: %w", domain.ErrInvalidInput)
	default:
		return fmt.Errorf("tipo de ocorrencia invalido: %w", domain.ErrInvalidInput)
	}
}
