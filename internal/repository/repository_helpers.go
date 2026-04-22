package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

const dateLayout = "2006-01-02"

func parseOptionalDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	return &parsed, nil
}

func parseRequiredDate(value string) (time.Time, error) {
	parsed, err := parseOptionalDate(value)
	if err != nil {
		return time.Time{}, err
	}
	if parsed == nil {
		return time.Time{}, domain.ErrInvalidInput
	}

	return *parsed, nil
}

func formatOptionalDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}

	return value.Format(dateLayout)
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
