package security

import (
	"errors"
	"os"
)

func DataEncryptionKeyFromEnv() (string, error) {
	if key := os.Getenv("DATA_ENCRYPTION_KEY"); key != "" {
		return key, nil
	}

	if fallback := os.Getenv("JWT_ACCESS_SECRET"); fallback != "" {
		return fallback, nil
	}

	return "", errors.New("variavel DATA_ENCRYPTION_KEY ou JWT_ACCESS_SECRET e obrigatoria para criptografia de dados")
}
