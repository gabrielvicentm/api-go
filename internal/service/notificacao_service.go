package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
)

type NotificacaoService struct {
	repo             *repository.NotificacaoRepository
	mu               sync.Mutex
	nextSubscriberID int64
	subscribers      map[int64]notificacaoSubscriber
}

func NewNotificacaoService(repo *repository.NotificacaoRepository) *NotificacaoService {
	return &NotificacaoService{
		repo:        repo,
		subscribers: make(map[int64]notificacaoSubscriber),
	}
}

func (s *NotificacaoService) Create(ctx context.Context, input domain.NotificacaoCreateRequest) (*domain.NotificacaoDetail, error) {
	normalized, err := normalizeNotificacaoCreate(input)
	if err != nil {
		return nil, err
	}

	item, err := s.repo.Create(ctx, normalized)
	if err != nil {
		return nil, err
	}

	s.broadcast(item)

	return item, nil
}

func (s *NotificacaoService) ListByRecipient(ctx context.Context, filter domain.NotificacaoListFilter) ([]domain.NotificacaoDetail, int64, error) {
	filter.DestinatarioTipo = strings.TrimSpace(strings.ToLower(filter.DestinatarioTipo))
	filter.DestinatarioID = strings.TrimSpace(filter.DestinatarioID)

	if err := validateNotificacaoRecipient(filter.DestinatarioTipo, filter.DestinatarioID); err != nil {
		return nil, 0, err
	}

	return s.repo.ListByRecipient(ctx, filter)
}

func (s *NotificacaoService) MarkAsRead(ctx context.Context, id, destinatarioTipo, destinatarioID string) (*domain.NotificacaoDetail, error) {
	id = strings.TrimSpace(id)
	destinatarioTipo = strings.TrimSpace(strings.ToLower(destinatarioTipo))
	destinatarioID = strings.TrimSpace(destinatarioID)

	if id == "" {
		return nil, fmt.Errorf("id da notificacao obrigatorio: %w", domain.ErrInvalidInput)
	}
	if err := validateNotificacaoRecipient(destinatarioTipo, destinatarioID); err != nil {
		return nil, err
	}

	return s.repo.MarkAsReadByRecipient(ctx, id, destinatarioTipo, destinatarioID)
}

type NotificacaoSubscription struct {
	Events <-chan domain.NotificacaoDetail
	close  func()
}

func (s *NotificacaoService) Subscribe(destinatarioTipo, destinatarioID string) (*NotificacaoSubscription, error) {
	destinatarioTipo = strings.TrimSpace(strings.ToLower(destinatarioTipo))
	destinatarioID = strings.TrimSpace(destinatarioID)

	if err := validateNotificacaoRecipient(destinatarioTipo, destinatarioID); err != nil {
		return nil, err
	}

	events := make(chan domain.NotificacaoDetail, 10)

	s.mu.Lock()
	s.nextSubscriberID++
	subscriberID := s.nextSubscriberID
	s.subscribers[subscriberID] = notificacaoSubscriber{
		destinatarioTipo: destinatarioTipo,
		destinatarioID:   destinatarioID,
		events:           events,
	}
	s.mu.Unlock()

	return &NotificacaoSubscription{
		Events: events,
		close: func() {
			s.mu.Lock()
			defer s.mu.Unlock()

			if subscriber, ok := s.subscribers[subscriberID]; ok {
				delete(s.subscribers, subscriberID)
				close(subscriber.events)
			}
		},
	}, nil
}

func (s *NotificacaoSubscription) Close() {
	if s == nil || s.close == nil {
		return
	}

	s.close()
}

type notificacaoSubscriber struct {
	destinatarioTipo string
	destinatarioID   string
	events           chan domain.NotificacaoDetail
}

func (s *NotificacaoService) broadcast(item *domain.NotificacaoDetail) {
	if item == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, subscriber := range s.subscribers {
		if !canReceiveNotificacao(subscriber, item) {
			continue
		}

		select {
		case subscriber.events <- *item:
		default:
		}
	}
}

func canReceiveNotificacao(subscriber notificacaoSubscriber, item *domain.NotificacaoDetail) bool {
	switch subscriber.destinatarioTipo {
	case domain.DestinatarioTipoAdmin:
		return item.DestinatarioTipo == "" ||
			(item.DestinatarioTipo == domain.DestinatarioTipoAdmin &&
				(item.DestinatarioID == "" || item.DestinatarioID == subscriber.destinatarioID))
	case domain.DestinatarioTipoMotorista:
		return item.DestinatarioTipo == domain.DestinatarioTipoMotorista &&
			item.DestinatarioID == subscriber.destinatarioID
	default:
		return false
	}
}

func normalizeNotificacaoCreate(input domain.NotificacaoCreateRequest) (domain.NotificacaoCreateRequest, error) {
	input.DestinatarioTipo = strings.TrimSpace(strings.ToLower(input.DestinatarioTipo))
	input.DestinatarioID = strings.TrimSpace(input.DestinatarioID)
	input.OrigemTipo = strings.TrimSpace(strings.ToLower(input.OrigemTipo))
	input.OrigemID = strings.TrimSpace(input.OrigemID)
	input.Titulo = strings.TrimSpace(input.Titulo)
	input.Mensagem = strings.TrimSpace(input.Mensagem)
	input.ReferenciaTipo = strings.TrimSpace(strings.ToLower(input.ReferenciaTipo))
	input.ReferenciaID = strings.TrimSpace(input.ReferenciaID)

	if input.Titulo == "" {
		return input, fmt.Errorf("titulo da notificacao obrigatorio: %w", domain.ErrInvalidInput)
	}

	switch input.DestinatarioTipo {
	case "", domain.DestinatarioTipoAdmin, domain.DestinatarioTipoMotorista:
	default:
		return input, fmt.Errorf("destinatario_tipo invalido: %w", domain.ErrInvalidInput)
	}

	if input.DestinatarioTipo == "" && input.DestinatarioID != "" {
		return input, fmt.Errorf("destinatario_tipo obrigatorio quando destinatario_id for informado: %w", domain.ErrInvalidInput)
	}

	if input.DestinatarioTipo == domain.DestinatarioTipoMotorista && input.DestinatarioID == "" {
		return input, fmt.Errorf("destinatario_id obrigatorio para notificacao de motorista: %w", domain.ErrInvalidInput)
	}

	switch input.OrigemTipo {
	case "", domain.OrigemTipoMotorista, domain.OrigemTipoSistema:
	default:
		return input, fmt.Errorf("origem_tipo invalido: %w", domain.ErrInvalidInput)
	}

	if input.OrigemTipo == "" && input.OrigemID != "" {
		return input, fmt.Errorf("origem_tipo obrigatorio quando origem_id for informado: %w", domain.ErrInvalidInput)
	}

	if input.ReferenciaTipo == "" && input.ReferenciaID != "" {
		return input, fmt.Errorf("referencia_tipo obrigatorio quando referencia_id for informado: %w", domain.ErrInvalidInput)
	}

	return input, nil
}

func validateNotificacaoRecipient(destinatarioTipo, destinatarioID string) error {
	switch destinatarioTipo {
	case domain.DestinatarioTipoAdmin:
		if destinatarioID == "" {
			return fmt.Errorf("destinatario_id obrigatorio para notificacoes de admin autenticado: %w", domain.ErrInvalidInput)
		}
		return nil
	case domain.DestinatarioTipoMotorista:
		if destinatarioID == "" {
			return fmt.Errorf("destinatario_id obrigatorio para notificacoes de motorista: %w", domain.ErrInvalidInput)
		}
		return nil
	default:
		return fmt.Errorf("destinatario_tipo invalido: %w", domain.ErrInvalidInput)
	}
}
