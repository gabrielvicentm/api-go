package service

import (
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
)

type AbastecimentoPhotoStorage interface {
	UploadAbastecimentoPhoto(ctx context.Context, body io.Reader, originalFilename, contentType string) (string, error)
}

type AbastecimentoService struct {
	repo         *repository.AbastecimentoRepository
	viagemRepo   *repository.ViagemRepository
	photoStorage AbastecimentoPhotoStorage
}

func NewAbastecimentoService(repo *repository.AbastecimentoRepository, viagemRepo *repository.ViagemRepository, photoStorage AbastecimentoPhotoStorage) *AbastecimentoService {
	return &AbastecimentoService{
		repo:         repo,
		viagemRepo:   viagemRepo,
		photoStorage: photoStorage,
	}
}

func (s *AbastecimentoService) Create(ctx context.Context, motoristaID string, input domain.AbastecimentoCreateRequest, photo io.Reader, photoFilename, photoContentType string) (*domain.AbastecimentoItem, error) {
	input.KMAtual = strings.TrimSpace(input.KMAtual)
	input.Litros = strings.TrimSpace(input.Litros)
	input.ValorPorLitro = strings.TrimSpace(input.ValorPorLitro)
	input.ValorTotal = strings.TrimSpace(input.ValorTotal)
	input.Fornecedor = strings.TrimSpace(input.Fornecedor)

	if _, err := parsePositiveDecimal(input.KMAtual); err != nil {
		return nil, fmt.Errorf("km atual invalido: %w", err)
	}
	litros, err := parsePositiveDecimal(input.Litros)
	if err != nil {
		return nil, fmt.Errorf("quantidade abastecida invalida: %w", err)
	}
	valorPorLitro, err := parsePositiveDecimal(input.ValorPorLitro)
	if err != nil {
		return nil, fmt.Errorf("valor do litro invalido: %w", err)
	}
	if input.ValorTotal != "" {
		valorTotal, err := parsePositiveDecimal(input.ValorTotal)
		if err != nil {
			return nil, fmt.Errorf("valor total invalido: %w", err)
		}
		if math.Abs(valorTotal-(litros*valorPorLitro)) > 0.01 {
			return nil, fmt.Errorf("valor total nao confere com litros x valor do litro: %w", domain.ErrInvalidInput)
		}
	}

	trip, err := s.currentMotoristaTrip(ctx, motoristaID)
	if err != nil {
		return nil, err
	}
	if trip.Status != "em_andamento" && trip.Status != "parada" {
		return nil, fmt.Errorf("viagem precisa estar em andamento ou parada para registrar abastecimento: %w", domain.ErrInvalidInput)
	}

	photoURL := ""
	if photo != nil {
		if s.photoStorage == nil {
			return nil, fmt.Errorf("armazenamento de fotos nao configurado: %w", domain.ErrInvalidInput)
		}

		uploadedURL, err := s.photoStorage.UploadAbastecimentoPhoto(ctx, photo, photoFilename, photoContentType)
		if err != nil {
			return nil, err
		}
		photoURL = uploadedURL
	}

	item, err := s.repo.Create(ctx, domain.AbastecimentoCreateInput{
		ViagemID:      trip.ID,
		VeiculoID:     trip.VeiculoID,
		MotoristaID:   motoristaID,
		KMAtual:       input.KMAtual,
		Litros:        input.Litros,
		ValorPorLitro: input.ValorPorLitro,
		Fornecedor:    input.Fornecedor,
		FotoURL:       photoURL,
	})
	if err != nil {
		return nil, err
	}

	if err := s.viagemRepo.CreateHistory(ctx, domain.ViagemHistoricoCreateInput{
		ViagemID:    trip.ID,
		UsuarioTipo: domain.ActorTypeMotorista,
		UsuarioID:   motoristaID,
		Descricao:   "Motorista registrou abastecimento",
	}); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *AbastecimentoService) List(ctx context.Context, filter domain.AbastecimentoListFilter) ([]domain.AbastecimentoItem, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *AbastecimentoService) GetByID(ctx context.Context, id string) (*domain.AbastecimentoItem, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AbastecimentoService) currentMotoristaTrip(ctx context.Context, motoristaID string) (*domain.ViagemDetail, error) {
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

func parsePositiveDecimal(value string) (float64, error) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed <= 0 {
		return 0, domain.ErrInvalidInput
	}

	return parsed, nil
}
