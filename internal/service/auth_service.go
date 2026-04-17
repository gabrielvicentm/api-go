package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gabrielvicentm/api-go.git/internal/security"
	"golang.org/x/crypto/bcrypt"
)

var nonDigitRegex = regexp.MustCompile(`\D`)

type AuthService struct {
	repo         domain.AuthRepository
	tokenManager *security.TokenManager
}

func NewAuthService(repo domain.AuthRepository, tokenManager *security.TokenManager) *AuthService {
	return &AuthService{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

func (s *AuthService) LoginAdmin(ctx context.Context, input domain.AdminLoginRequest) (*domain.TokenResponse, error) {
	actor, err := s.repo.FindAdminByEmail(ctx, normalizeEmail(input.Email))
	if err != nil {
		return nil, err
	}

	return s.authenticateAndIssueTokens(ctx, actor, input.Senha)
}

func (s *AuthService) LoginMotorista(ctx context.Context, input domain.MotoristaLoginRequest) (*domain.TokenResponse, error) {
	cpf := normalizeCPF(input.CPF)
	if len(cpf) != 11 {
		return nil, domain.ErrInvalidCredentials
	}

	actor, err := s.repo.FindMotoristaByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}

	return s.authenticateAndIssueTokens(ctx, actor, input.Senha)
}

func (s *AuthService) RefreshToken(ctx context.Context, input domain.RefreshTokenRequest) (*domain.TokenResponse, error) {
	claims, err := s.tokenManager.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, err
	}

	session, err := s.repo.FindRefreshSessionByTokenID(ctx, claims.TokenID)
	if err != nil {
		return nil, err
	}

	if session.ActorID != claims.UserID || session.ActorType != claims.ActorType {
		return nil, domain.ErrInvalidToken
	}

	if session.RevokedAt != nil {
		return nil, domain.ErrInvalidToken
	}

	if session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, domain.ErrExpiredToken
	}

	if session.TokenHash != s.tokenManager.HashRefreshToken(input.RefreshToken) {
		return nil, domain.ErrInvalidToken
	}

	actor, err := s.repo.FindActorByID(ctx, claims.ActorType, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, domain.ErrInvalidToken
		}

		return nil, err
	}

	if !actor.Ativo {
		return nil, domain.ErrInactiveUser
	}

	if err := s.repo.RevokeRefreshSession(ctx, claims.TokenID); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateLastAccess(ctx, actor.ActorType, actor.ID); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, actor)
}

func (s *AuthService) Logout(ctx context.Context, input domain.LogoutRequest) error {
	claims, err := s.tokenManager.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return err
	}

	session, err := s.repo.FindRefreshSessionByTokenID(ctx, claims.TokenID)
	if err != nil {
		return err
	}

	if session.ActorID != claims.UserID || session.ActorType != claims.ActorType {
		return domain.ErrInvalidToken
	}

	if session.TokenHash != s.tokenManager.HashRefreshToken(input.RefreshToken) {
		return domain.ErrInvalidToken
	}

	return s.repo.RevokeRefreshSession(ctx, claims.TokenID)
}

func (s *AuthService) GetProfile(ctx context.Context, actorType, actorID string) (*domain.AuthUserResponse, error) {
	actor, err := s.repo.FindActorByID(ctx, actorType, actorID)
	if err != nil {
		return nil, err
	}

	if !actor.Ativo {
		return nil, domain.ErrInactiveUser
	}

	return buildUserResponse(actor), nil
}

func (s *AuthService) authenticateAndIssueTokens(ctx context.Context, actor *domain.AuthenticatedActor, senha string) (*domain.TokenResponse, error) {
	if !actor.Ativo {
		return nil, domain.ErrInactiveUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(actor.SenhaHash), []byte(senha)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, domain.ErrInvalidCredentials
		}

		return nil, err
	}

	if err := s.repo.UpdateLastAccess(ctx, actor.ActorType, actor.ID); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, actor)
}

func (s *AuthService) issueTokens(ctx context.Context, actor *domain.AuthenticatedActor) (*domain.TokenResponse, error) {
	tokenID, err := s.tokenManager.GenerateTokenID()
	if err != nil {
		return nil, err
	}

	accessToken, _, err := s.tokenManager.GenerateAccessToken(actor)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken(actor.ActorType, actor.ID, tokenID)
	if err != nil {
		return nil, err
	}

	err = s.repo.CreateRefreshSession(ctx, domain.RefreshSession{
		TokenID:   tokenID,
		ActorID:   actor.ID,
		ActorType: actor.ActorType,
		TokenHash: s.tokenManager.HashRefreshToken(refreshToken),
		ExpiresAt: refreshExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	return &domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.tokenManager.AccessTTLSeconds(),
		User:         *buildUserResponse(actor),
	}, nil
}

func buildUserResponse(actor *domain.AuthenticatedActor) *domain.AuthUserResponse {
	return &domain.AuthUserResponse{
		ID:        actor.ID,
		Nome:      actor.Nome,
		Email:     actor.Email,
		Role:      actor.Role,
		ActorType: actor.ActorType,
	}
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func normalizeCPF(cpf string) string {
	return nonDigitRegex.ReplaceAllString(cpf, "")
}
