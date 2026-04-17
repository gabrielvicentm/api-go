package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(".env", "../../.env"); err != nil {

		log.Fatal("Erro ao carregar .env: ", err)

	}
}

func NewDBConnection() *pgxpool.Pool {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",

		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	pool, err := pgxpool.New(context.Background(), dsn)

	if err != nil {
		log.Fatal("Erro ao conectar no banco: ", err)

	}

	return pool
}
