package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
)

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	refreshPepper string
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type accessPayload struct {
	Subject   string `json:"sub"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	ActorType string `json:"actor_type"`
	Type      string `json:"type"`
	Issuer    string `json:"iss"`
	Issued    int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	Expires   int64  `json:"exp"`
}

type refreshPayload struct {
	Subject   string `json:"sub"`
	TokenID   string `json:"jti"`
	ActorType string `json:"actor_type"`
	Type      string `json:"type"`
	Issuer    string `json:"iss"`
	Issued    int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	Expires   int64  `json:"exp"`
}

func NewTokenManagerFromEnv() (*TokenManager, error) {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	refreshPepper := os.Getenv("REFRESH_TOKEN_PEPPER")

	if accessSecret == "" || refreshSecret == "" || refreshPepper == "" {
		return nil, errors.New("variaveis JWT_ACCESS_SECRET, JWT_REFRESH_SECRET e REFRESH_TOKEN_PEPPER sao obrigatorias")
	}

	accessTTL, err := parseDurationFromEnv("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshTTL, err := parseDurationFromEnv("JWT_REFRESH_TTL", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "api-go"
	}

	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		refreshPepper: refreshPepper,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		issuer:        issuer,
	}, nil
}

func (tm *TokenManager) GenerateTokenID() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}

func (tm *TokenManager) GenerateAccessToken(actor *domain.AuthenticatedActor) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(tm.accessTTL)

	payload := accessPayload{
		Subject:   actor.ID,
		Email:     actor.Email,
		Role:      actor.Role,
		ActorType: actor.ActorType,
		Type:      "access",
		Issuer:    tm.issuer,
		Issued:    now.Unix(),
		NotBefore: now.Unix(),
		Expires:   expiresAt.Unix(),
	}

	token, err := tm.signJWT(payload, tm.accessSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (tm *TokenManager) GenerateRefreshToken(actorType, actorID, tokenID string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(tm.refreshTTL)

	payload := refreshPayload{
		Subject:   actorID,
		TokenID:   tokenID,
		ActorType: actorType,
		Type:      "refresh",
		Issuer:    tm.issuer,
		Issued:    now.Unix(),
		NotBefore: now.Unix(),
		Expires:   expiresAt.Unix(),
	}

	token, err := tm.signJWT(payload, tm.refreshSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (tm *TokenManager) ValidateAccessToken(token string) (*domain.AccessTokenClaims, error) {
	payloadBytes, err := tm.verifyJWT(token, tm.accessSecret)
	if err != nil {
		return nil, err
	}

	var payload accessPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, domain.ErrInvalidToken
	}

	if payload.Type != "access" || payload.Subject == "" || payload.Role == "" || payload.ActorType == "" {
		return nil, domain.ErrInvalidToken
	}

	return &domain.AccessTokenClaims{
		UserID:    payload.Subject,
		Email:     payload.Email,
		Role:      payload.Role,
		ActorType: payload.ActorType,
	}, nil
}

func (tm *TokenManager) ValidateRefreshToken(token string) (*domain.RefreshTokenClaims, error) {
	payloadBytes, err := tm.verifyJWT(token, tm.refreshSecret)
	if err != nil {
		return nil, err
	}

	var payload refreshPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, domain.ErrInvalidToken
	}

	if payload.Type != "refresh" || payload.Subject == "" || payload.TokenID == "" || payload.ActorType == "" {
		return nil, domain.ErrInvalidToken
	}

	return &domain.RefreshTokenClaims{
		TokenID:   payload.TokenID,
		UserID:    payload.Subject,
		ActorType: payload.ActorType,
	}, nil
}

func (tm *TokenManager) HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(tm.refreshPepper + ":" + token))
	return hex.EncodeToString(sum[:])
}

func (tm *TokenManager) AccessTTLSeconds() int64 {
	return int64(tm.accessTTL.Seconds())
}

func (tm *TokenManager) signJWT(payload any, secret []byte) (string, error) {
	headerBytes, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadBytes)
	unsignedToken := headerPart + "." + payloadPart

	signature := tm.sign(unsignedToken, secret)

	return unsignedToken + "." + signature, nil
}

func (tm *TokenManager) verifyJWT(token string, secret []byte) ([]byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, domain.ErrInvalidToken
	}

	unsignedToken := parts[0] + "." + parts[1]
	expectedSignature := tm.sign(unsignedToken, secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return nil, domain.ErrInvalidToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, domain.ErrInvalidToken
	}

	if header.Alg != "HS256" || header.Typ != "JWT" {
		return nil, domain.ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	var registered struct {
		Issuer    string `json:"iss"`
		Issued    int64  `json:"iat"`
		NotBefore int64  `json:"nbf"`
		Expires   int64  `json:"exp"`
	}

	if err := json.Unmarshal(payloadBytes, &registered); err != nil {
		return nil, domain.ErrInvalidToken
	}

	now := time.Now().UTC().Unix()

	if registered.Issuer != tm.issuer || registered.Issued == 0 || registered.Expires == 0 {
		return nil, domain.ErrInvalidToken
	}

	if registered.NotBefore > now {
		return nil, domain.ErrInvalidToken
	}

	if registered.Expires <= now {
		return nil, domain.ErrExpiredToken
	}

	return payloadBytes, nil
}

func (tm *TokenManager) sign(value string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func parseDurationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("valor invalido para %s", key)
	}

	return duration, nil
}
