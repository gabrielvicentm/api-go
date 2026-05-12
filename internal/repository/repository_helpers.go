package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const dateLayout = "2006-01-02"
const timeLayout = "15:04"
const localSeedEncryptionKey = "dev-local-key"

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func parseOptionalDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return nil, fmt.Errorf("data invalida, use o formato YYYY-MM-DD: %w", domain.ErrInvalidInput)
	}

	return &parsed, nil
}

func parseRequiredDate(value string) (time.Time, error) {
	parsed, err := parseOptionalDate(value)
	if err != nil {
		return time.Time{}, err
	}
	if parsed == nil {
		return time.Time{}, fmt.Errorf("data obrigatoria: %w", domain.ErrInvalidInput)
	}

	return *parsed, nil
}

func formatOptionalDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}

	return value.Format(dateLayout)
}

func parseOptionalTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(timeLayout, value)
	if err != nil {
		return nil, fmt.Errorf("horario invalido, use o formato HH:MM: %w", domain.ErrInvalidInput)
	}

	return &parsed, nil
}

func formatOptionalTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}

	return value.Format(timeLayout)
}

func normalizeNullableEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func normalizeDigits(value string) string {
	replacer := strings.NewReplacer(".", "", "-", "", "/", "", "(", "", ")", "", " ", "")
	return replacer.Replace(strings.TrimSpace(value))
}

func maskCPF(cpf string) string {
	if len(cpf) != 11 {
		return cpf
	}
	return cpf[:3] + "." + cpf[3:6] + "." + cpf[6:9] + "-" + cpf[9:]
}

func decryptTextField(ctx context.Context, db queryRower, encrypted []byte, primaryKey string) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}

	keys := []string{strings.TrimSpace(primaryKey)}
	if localSeedEncryptionKey != strings.TrimSpace(primaryKey) {
		keys = append(keys, localSeedEncryptionKey)
	}

	var lastErr error
	for _, key := range keys {
		if key == "" {
			continue
		}

		var value string
		err := db.QueryRow(ctx, `SELECT pgp_sym_decrypt($1::bytea, $2)::text`, encrypted, key).Scan(&value)
		if err == nil {
			return value, nil
		}

		lastErr = err
	}

	if lastErr == nil {
		return "", domain.ErrInvalidInput
	}

	return "", lastErr
}

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return domain.ErrConflict
		case "23503":
			return domain.ErrConflict
		case "22P02", "23514":
			return domain.ErrInvalidInput
		}
	}

	return err
}
