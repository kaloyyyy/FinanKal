package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type LedgerRepository struct {
	db *pgxpool.Pool
}

func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) CreateTransaction(ctx context.Context, description string) (uuid.UUID, error) {
	var txID uuid.UUID

	err := r.db.QueryRow(ctx,
		"INSERT INTO transactions (description) VALUES ($1) RETURNING id",
		description,
	).Scan(&txID)

	return txID, err
}

func (r *LedgerRepository) InsertEntry(ctx context.Context, txID uuid.UUID, accountID uuid.UUID, amount decimal.Decimal,
	entryType string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO entries (transaction_id, account_id, amount, type)
		 VALUES ($1, $2, $3, $4)`,
		txID, accountID, amount, entryType,
	)

	return err
}

func (r *LedgerRepository) GetBalance(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	var balance decimal.Decimal

	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(
			SUM(CASE WHEN type='DEBIT' THEN amount ELSE 0 END) -
			SUM(CASE WHEN type='CREDIT' THEN amount ELSE 0 END), 0)
		FROM entries
		WHERE account_id = $1
	`, accountID).Scan(&balance)

	return balance, err
}

func (r *LedgerRepository) GetAccount(ctx context.Context, accountID uuid.UUID) (string, string, time.Time, error) {
	var name, accountType string

	var createAt time.Time
	err := r.db.QueryRow(ctx,
		`SELECT name, type, created_at FROM accounts WHERE id = $1`,
		accountID,
	).Scan(&name, &accountType, &createAt)

	return name, accountType, createAt, err
}
