package repository

import (
	"context"
	"errors"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindAdminByEmail(ctx context.Context, email string) (*domain.AuthenticatedActor, error) {
	const query = `
		SELECT id, nome, email, senha_hash, role, ativo
		FROM usuarios
		WHERE email = $1
		LIMIT 1
	`

	actor, err := r.scanAdmin(ctx, query, email)
	if err != nil {
		return nil, err
	}

	actor.ActorType = domain.ActorTypeAdmin
	return actor, nil
}

func (r *AuthRepository) FindMotoristaByCPF(ctx context.Context, cpf string) (*domain.AuthenticatedActor, error) {
	const query = `
		SELECT
			m.id,
			m.nome,
			COALESCE(m.email, ''),
			mc.senha_hash,
			m.status = 'ativo' AND mc.ativo AS ativo
		FROM motoristas m
		JOIN motorista_credenciais mc ON mc.motorista_id = m.id
		WHERE m.cpf_hash = encode(digest($1, 'sha256'), 'hex')
		LIMIT 1
	`

	var actor domain.AuthenticatedActor

	err := r.db.QueryRow(ctx, query, cpf).Scan(
		&actor.ID,
		&actor.Nome,
		&actor.Email,
		&actor.SenhaHash,
		&actor.Ativo,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}

		return nil, err
	}

	actor.ActorType = domain.ActorTypeMotorista
	actor.Role = domain.RoleMotorista
	actor.DocumentID = cpf

	return &actor, nil
}

func (r *AuthRepository) FindActorByID(ctx context.Context, actorType, actorID string) (*domain.AuthenticatedActor, error) {
	switch actorType {
	case domain.ActorTypeAdmin:
		const query = `
			SELECT id, nome, email, senha_hash, role, ativo
			FROM usuarios
			WHERE id = $1
			LIMIT 1
		`

		actor, err := r.scanAdmin(ctx, query, actorID)
		if err != nil {
			return nil, err
		}

		actor.ActorType = domain.ActorTypeAdmin
		return actor, nil
	case domain.ActorTypeMotorista:
		const query = `
			SELECT
				m.id,
				m.nome,
				COALESCE(m.email, ''),
				mc.senha_hash,
				m.status = 'ativo' AND mc.ativo AS ativo
			FROM motoristas m
			JOIN motorista_credenciais mc ON mc.motorista_id = m.id
			WHERE m.id = $1
			LIMIT 1
		`

		var actor domain.AuthenticatedActor

		err := r.db.QueryRow(ctx, query, actorID).Scan(
			&actor.ID,
			&actor.Nome,
			&actor.Email,
			&actor.SenhaHash,
			&actor.Ativo,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrInvalidToken
			}

			return nil, err
		}

		actor.ActorType = domain.ActorTypeMotorista
		actor.Role = domain.RoleMotorista
		return &actor, nil
	default:
		return nil, domain.ErrInvalidToken
	}
}

func (r *AuthRepository) UpdateLastAccess(ctx context.Context, actorType, actorID string) error {
	switch actorType {
	case domain.ActorTypeAdmin:
		const query = `
			UPDATE usuarios
			SET ultimo_acesso = NOW(), updated_at = NOW()
			WHERE id = $1
		`

		_, err := r.db.Exec(ctx, query, actorID)
		return err
	case domain.ActorTypeMotorista:
		const query = `
			UPDATE motorista_credenciais
			SET ultimo_acesso = NOW(), updated_at = NOW()
			WHERE motorista_id = $1
		`

		_, err := r.db.Exec(ctx, query, actorID)
		return err
	default:
		return domain.ErrInvalidToken
	}
}

func (r *AuthRepository) CreateRefreshSession(ctx context.Context, session domain.RefreshSession) error {
	const query = `
		INSERT INTO auth_refresh_tokens (token_id, actor_id, actor_type, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query, session.TokenID, session.ActorID, session.ActorType, session.TokenHash, session.ExpiresAt)
	return err
}

func (r *AuthRepository) FindRefreshSessionByTokenID(ctx context.Context, tokenID string) (*domain.RefreshSession, error) {
	const query = `
		SELECT token_id, actor_id, actor_type, token_hash, expires_at, revoked_at
		FROM auth_refresh_tokens
		WHERE token_id = $1
		LIMIT 1
	`

	var session domain.RefreshSession

	err := r.db.QueryRow(ctx, query, tokenID).Scan(
		&session.TokenID,
		&session.ActorID,
		&session.ActorType,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}

		return nil, err
	}

	return &session, nil
}

func (r *AuthRepository) RevokeRefreshSession(ctx context.Context, tokenID string) error {
	const query = `
		UPDATE auth_refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE token_id = $1 AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, tokenID)
	return err
}

func (r *AuthRepository) scanAdmin(ctx context.Context, query string, value string) (*domain.AuthenticatedActor, error) {
	var actor domain.AuthenticatedActor

	err := r.db.QueryRow(ctx, query, value).Scan(
		&actor.ID,
		&actor.Nome,
		&actor.Email,
		&actor.SenhaHash,
		&actor.Role,
		&actor.Ativo,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}

		return nil, err
	}

	return &actor, nil
}
