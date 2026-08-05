package ledger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type LedgerRepository struct {
	db *pgxpool.Pool
}

func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) CreateTransaction(ctx context.Context, userID uuid.UUID, description string) (uuid.UUID, error) {
	var txID uuid.UUID

	err := r.db.QueryRow(ctx,
		"INSERT INTO transactions (user_id, description) VALUES ($1, $2) RETURNING id",
		userID, description,
	).Scan(&txID)

	return txID, err
}

func (r *LedgerRepository) InsertEntry(
	ctx context.Context,
	txID uuid.UUID,
	accountID uuid.UUID,
	amount decimal.Decimal,
	entryType EntryType,
) (uuid.UUID, error) {

	var id uuid.UUID

	const query = `
INSERT INTO entries (
	transaction_id,
	account_id,
	amount,
	type
)
VALUES (
	$1,
	$2,
	$3,
	$4
)
RETURNING id
`

	err := r.db.QueryRow(
		ctx,
		query,
		txID,
		accountID,
		amount,
		entryType,
	).Scan(&id)

	return id, err
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

func (r *LedgerRepository) GetAccount(ctx context.Context, accountID uuid.UUID) (string, AccountType, time.Time, error) {
	var name string
	var accountType AccountType
	var createAt time.Time
	err := r.db.QueryRow(ctx,
		`SELECT name, type, created_at FROM accounts WHERE id = $1`,
		accountID,
	).Scan(&name, &accountType, &createAt)

	return name, accountType, createAt, err
}

func (r *LedgerRepository) GetLedgerEntries(ctx context.Context, accountID uuid.UUID) ([]Entry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT e.transaction_id, e.account_id, e.amount, e.type, t.created_at, t.description
		FROM entries e
		JOIN transactions t ON e.transaction_id = t.id
		WHERE e.account_id = $1
		ORDER BY t.created_at DESC
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		var entryType string
		err := rows.Scan(&entry.TransactionID, &entry.AccountID, &entry.Amount, &entryType, &entry.CreatedAt, &entry.Description)
		if err != nil {
			return nil, err
		}
		entry.Type = EntryType(entryType)
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (r *LedgerRepository) GetUserIDFromAccount(ctx context.Context, accountID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.QueryRow(ctx,
		`SELECT user_id FROM accounts WHERE id = $1`,
		accountID,
	).Scan(&userID)
	return userID, err
}

func (r *LedgerRepository) GetUserTotalCredit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM entries e
		JOIN accounts a ON e.account_id = a.id
		WHERE a.user_id = $1 AND e.type = 'CREDIT'
	`, userID).Scan(&total)
	return total, err
}

func (r *LedgerRepository) GetUserTotalDebit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM entries e
		JOIN accounts a ON e.account_id = a.id
		WHERE a.user_id = $1 AND e.type = 'DEBIT'
	`, userID).Scan(&total)
	return total, err
}

func (r *LedgerRepository) GetUserTotalBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(CASE WHEN e.type = 'DEBIT' THEN amount ELSE -amount END), 0)
		FROM entries e
		JOIN accounts a ON e.account_id = a.id
		WHERE a.user_id = $1
	`, userID).Scan(&total)
	return total, err
}

func (r *LedgerRepository) BeginTx(
	ctx context.Context,
) (pgx.Tx, error) {

	return r.db.Begin(ctx)
}

func (r *LedgerRepository) CreateTransactionTx(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
	description string,
) (uuid.UUID, error) {

	var transactionID uuid.UUID

	err := tx.QueryRow(
		ctx,
		`
		INSERT INTO transactions (
			user_id,
			description
		)
		VALUES ($1,$2)
		RETURNING id
		`,
		userID,
		description,
	).Scan(&transactionID)

	return transactionID, err
}

func (r *LedgerRepository) InsertEntryTx(
	ctx context.Context,
	tx pgx.Tx,
	transactionID uuid.UUID,
	accountID uuid.UUID,
	amount decimal.Decimal,
	entryType EntryType,
) (uuid.UUID, error) {

	var entryID uuid.UUID

	err := tx.QueryRow(
		ctx,
		`
		INSERT INTO entries (
			transaction_id,
			account_id,
			amount,
			type
		)
		VALUES ($1,$2,$3,$4)
		RETURNING id
		`,
		transactionID,
		accountID,
		amount,
		entryType,
	).Scan(&entryID)

	return entryID, err
}
