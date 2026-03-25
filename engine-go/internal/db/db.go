package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB() (*pgxpool.Pool, error) {
	println("Connecting to database...", os.Getenv("DATABASE_URL"))
	return pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
}
