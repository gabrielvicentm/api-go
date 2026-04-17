package domain

import (
	"context"
	"time"
)

const (
	ActorTypeAdmin     = "admin"
	ActorTypeMotorista = "motorista"

	RoleMotorista = "motorista"
)

type AdminLoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required,min=6"`
}

type MotoristaLoginRequest struct {
	CPF   string `json:"cpf" binding:"required"`
	Senha string `json:"senha" binding:"required,min=6"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	SenhaAtual string `json:"senha_atual" binding:"required,min=6"`
	NovaSenha  string `json:"nova_senha" binding:"required,min=6"`
}

type AuthenticatedActor struct {
	ID         string
	Nome       string
	Email      string
	SenhaHash  string
	Role       string
	ActorType  string
	Ativo      bool
	DocumentID string
}

type AuthUserResponse struct {
	ID        string `json:"id"`
	Nome      string `json:"nome"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	ActorType string `json:"actor_type"`
}

type TokenResponse struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	TokenType    string           `json:"token_type"`
	ExpiresIn    int64            `json:"expires_in"`
	User         AuthUserResponse `json:"user"`
}

type RefreshSession struct {
	TokenID   string
	ActorID   string
	ActorType string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type AccessTokenClaims struct {
	UserID    string
	Email     string
	Role      string
	ActorType string
}

type RefreshTokenClaims struct {
	TokenID   string
	UserID    string
	ActorType string
}

type AuthRepository interface {
	FindAdminByEmail(ctx context.Context, email string) (*AuthenticatedActor, error)
	FindMotoristaByCPF(ctx context.Context, cpf string) (*AuthenticatedActor, error)
	FindActorByID(ctx context.Context, actorType, actorID string) (*AuthenticatedActor, error)
	UpdateLastAccess(ctx context.Context, actorType, actorID string) error
	UpdatePassword(ctx context.Context, actorType, actorID, senhaHash string) error
	CreateRefreshSession(ctx context.Context, session RefreshSession) error
	FindRefreshSessionByTokenID(ctx context.Context, tokenID string) (*RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, tokenID string) error
	RevokeAllRefreshSessions(ctx context.Context, actorType, actorID string) error
}

type AuthService interface {
	LoginAdmin(ctx context.Context, input AdminLoginRequest) (*TokenResponse, error)
	LoginMotorista(ctx context.Context, input MotoristaLoginRequest) (*TokenResponse, error)
	RefreshToken(ctx context.Context, input RefreshTokenRequest) (*TokenResponse, error)
	Logout(ctx context.Context, input LogoutRequest) error
	ChangePassword(ctx context.Context, actorType, actorID string, input ChangePasswordRequest) error
	GetProfile(ctx context.Context, actorType, actorID string) (*AuthUserResponse, error)
}
