package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/repository"
)

type NotificacaoService struct {
	repo             *repository.NotificacaoRepository
	pushClient       PushSender
	mu               sync.Mutex
	nextSubscriberID int64
	subscribers      map[int64]notificacaoSubscriber
}

type PushSender interface {
	Send(ctx context.Context, messages []ExpoPushMessage) ([]ExpoPushTicket, error)
}

func NewNotificacaoService(repo *repository.NotificacaoRepository, pushClient PushSender) *NotificacaoService {
	return &NotificacaoService{
		repo:        repo,
		pushClient:  pushClient,
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
	s.sendPushAsync(item)

	return item, nil
}

func (s *NotificacaoService) RegisterPushToken(ctx context.Context, actorType, actorID string, input domain.PushTokenRegisterRequest) (*domain.PushTokenDetail, error) {
	normalized, err := normalizePushTokenRegister(actorType, actorID, input)
	if err != nil {
		return nil, err
	}

	return s.repo.UpsertPushToken(ctx, actorType, actorID, normalized)
}

func (s *NotificacaoService) DeactivatePushToken(ctx context.Context, actorType, actorID string, input domain.PushTokenDeleteRequest) error {
	actorType = strings.TrimSpace(strings.ToLower(actorType))
	actorID = strings.TrimSpace(actorID)
	token := strings.TrimSpace(input.Token)

	if err := validateActorRecipient(actorType, actorID); err != nil {
		return err
	}
	if !isExpoPushToken(token) {
		return fmt.Errorf("token expo invalido: %w", domain.ErrInvalidInput)
	}

	return s.repo.DeactivatePushToken(ctx, actorType, actorID, token)
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

func (s *NotificacaoService) sendPushAsync(item *domain.NotificacaoDetail) {
	if s.pushClient == nil || item == nil {
		return
	}

	if item.DestinatarioTipo == "" {
		s.sendPushForRecipientAsync(domain.DestinatarioTipoAdmin, "", item)
		return
	}

	s.sendPushForRecipientAsync(item.DestinatarioTipo, item.DestinatarioID, item)
}

func (s *NotificacaoService) sendPushForRecipientAsync(destinatarioTipo, destinatarioID string, item *domain.NotificacaoDetail) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		tokens, err := s.repo.ListPushTokensByRecipient(ctx, destinatarioTipo, destinatarioID)
		if err != nil {
			log.Printf("erro ao buscar tokens push: %v", err)
			return
		}
		if len(tokens) == 0 {
			return
		}

		messages := make([]ExpoPushMessage, 0, len(tokens))
		for _, token := range tokens {
			messages = append(messages, ExpoPushMessage{
				To:    token,
				Title: item.Titulo,
				Body:  item.Mensagem,
				Sound: "default",
				Data: map[string]any{
					"notificacao_id":    item.ID,
					"referencia_tipo":   item.ReferenciaTipo,
					"referencia_id":     item.ReferenciaID,
					"destinatario_tipo": item.DestinatarioTipo,
					"destinatario_id":   item.DestinatarioID,
				},
			})
		}

		tickets, err := s.pushClient.Send(ctx, messages)
		if err != nil {
			log.Printf("erro ao enviar push expo: %v", err)
			return
		}

		for _, ticket := range tickets {
			if ticket.Error != "DeviceNotRegistered" {
				continue
			}
			if err := s.repo.DeactivatePushTokenValue(ctx, ticket.Token); err != nil {
				log.Printf("erro ao desativar token push invalido: %v", err)
			}
		}
	}()
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

func normalizePushTokenRegister(actorType, actorID string, input domain.PushTokenRegisterRequest) (domain.PushTokenRegisterRequest, error) {
	actorType = strings.TrimSpace(strings.ToLower(actorType))
	actorID = strings.TrimSpace(actorID)
	input.Token = strings.TrimSpace(input.Token)
	input.Platform = strings.TrimSpace(strings.ToLower(input.Platform))
	input.DeviceID = strings.TrimSpace(input.DeviceID)

	if err := validateActorRecipient(actorType, actorID); err != nil {
		return input, err
	}

	if !isExpoPushToken(input.Token) {
		return input, fmt.Errorf("token expo invalido: %w", domain.ErrInvalidInput)
	}

	switch input.Platform {
	case "", "android", "ios", "web":
	default:
		return input, fmt.Errorf("platform invalida: %w", domain.ErrInvalidInput)
	}

	return input, nil
}

func validateActorRecipient(actorType, actorID string) error {
	switch actorType {
	case domain.ActorTypeAdmin, domain.ActorTypeMotorista:
		if actorID == "" {
			return fmt.Errorf("actor_id obrigatorio para push token: %w", domain.ErrInvalidInput)
		}
		return nil
	default:
		return fmt.Errorf("actor_type invalido para push token: %w", domain.ErrInvalidInput)
	}
}

func isExpoPushToken(token string) bool {
	return strings.HasPrefix(token, "ExpoPushToken[") || strings.HasPrefix(token, "ExponentPushToken[")
}
