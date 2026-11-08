package db

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const connectAttempts = 30

func NewDB() (*pgxpool.Pool, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	// pgxpool connects lazily, so without this a bad URL only fails on the first request.
	// Retries because docker compose starts this service before Postgres accepts connections.
	for attempt := 1; ; attempt++ {
		err := pool.Ping(ctx)
		if err == nil {
			return pool, nil
		}
		if attempt == connectAttempts {
			pool.Close()
			return nil, err
		}
		time.Sleep(time.Second)
	}
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
