package db

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB() (*pgxpool.Pool, error) {
	println("Connecting to database...", os.Getenv("DATABASE_URL"))
	return pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
}
func LogPgError(err error) {
	if err == nil {
		return
	}

	log.Printf("Error Type: %T", err)
	log.Printf("Error: %v", err)

	var prepareErr *pgconn.PrepareError
	if errors.As(err, &prepareErr) {
		log.Printf("PrepareError: %+v", prepareErr)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Printf("SQLSTATE: %s", pgErr.Code)
		log.Printf("Message : %s", pgErr.Message)
		log.Printf("Detail  : %s", pgErr.Detail)
		log.Printf("Hint    : %s", pgErr.Hint)
		log.Printf("Table   : %s", pgErr.TableName)
		log.Printf("Column  : %s", pgErr.ColumnName)
		log.Printf("Constraint: %s", pgErr.ConstraintName)
	}
}
