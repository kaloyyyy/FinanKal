package ledger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

const balanceQuery = `
	SELECT COALESCE(SUM(CASE WHEN type = 'DEBIT' THEN amount ELSE -amount END), 0)
	FROM entries
	WHERE account_id = $1
`

type LedgerRepository struct {
	db *pgxpool.Pool
}

func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) GetBalance(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	var balance decimal.Decimal

	err := r.db.QueryRow(ctx, balanceQuery, accountID).Scan(&balance)

	return balance, err
}

func (r *LedgerRepository) GetBalanceTx(ctx context.Context, tx pgx.Tx, accountID uuid.UUID) (decimal.Decimal, error) {
	var balance decimal.Decimal

	err := tx.QueryRow(ctx, balanceQuery, accountID).Scan(&balance)

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

func (r *LedgerRepository) CreateAccountTx(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
	name string,
	accountType AccountType,
) (uuid.UUID, error) {

	var accountID uuid.UUID

	err := tx.QueryRow(
		ctx,
		`
		INSERT INTO accounts (
			user_id,
			name,
			type
		)
		VALUES ($1,$2,$3)
		RETURNING id
		`,
		userID,
		name,
		accountType,
	).Scan(&accountID)

	return accountID, err
}

func (r *LedgerRepository) CreateClearingAccountTx(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
	cardName string,
) (uuid.UUID, error) {

	return r.CreateAccountTx(ctx, tx, userID, cardName+" Clearing", CLEARING)
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

func (r *LedgerRepository) GetUserTotalCredit(
	ctx context.Context,
	userID uuid.UUID,
) (decimal.Decimal, error) {

	var total decimal.Decimal

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COALESCE(
			SUM(
				CASE
					WHEN e.type = 'CREDIT'
						THEN e.amount
					ELSE
						-e.amount
				END
			),
			0
		)
		FROM entries e
		JOIN accounts a
			ON a.id = e.account_id
		WHERE a.user_id = $1
		AND a.type IN (
			'LIABILITY',
			'CREDIT_CARD'
		)
		`,
		userID,
	).Scan(&total)

	return total, err
}

func (r *LedgerRepository) GetUserTotalDebit(
	ctx context.Context,
	userID uuid.UUID,
) (decimal.Decimal, error) {

	var total decimal.Decimal

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COALESCE(
			SUM(
				CASE
					WHEN e.type = 'DEBIT'
						THEN e.amount
					ELSE
						-e.amount
				END
			),
			0
		)
		FROM entries e
		JOIN accounts a
			ON a.id = e.account_id
		WHERE a.user_id = $1
		AND a.type = 'ASSET'
		`,
		userID,
	).Scan(&total)

	return total, err
}

func (r *LedgerRepository) GetUserNetWorth(
	ctx context.Context,
	userID uuid.UUID,
) (decimal.Decimal, error) {

	var total decimal.Decimal

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COALESCE(
			SUM(
				CASE

					-- Assets
					WHEN a.type = 'ASSET' THEN
						CASE
							WHEN e.type = 'DEBIT'
								THEN e.amount
							ELSE
								-e.amount
						END

					-- Liabilities
					WHEN a.type IN (
						'LIABILITY',
						'CREDIT_CARD'
					) THEN
						CASE
							WHEN e.type = 'CREDIT'
								THEN -e.amount
							ELSE
								e.amount
						END

					ELSE 0

				END
			),
			0
		)
		FROM entries e
		JOIN accounts a
			ON a.id = e.account_id
		WHERE a.user_id = $1
		`,
		userID,
	).Scan(&total)

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
	occurredAt *time.Time, // nil = now
) (uuid.UUID, error) {

	var transactionID uuid.UUID

	err := tx.QueryRow(
		ctx,
		`
		INSERT INTO transactions (
			user_id,
			description,
			created_at
		)
		VALUES ($1,$2,COALESCE($3::timestamptz, now()))
		RETURNING id
		`,
		userID,
		description,
		occurredAt,
	).Scan(&transactionID)

	return transactionID, err
}

// DeleteUserTransactionsTx deletes every transaction touching the user's accounts (entries and
// statement links cascade) plus the user's credit card and bank statements, which are built from
// those transactions (and would otherwise block re-importing). Accounts and credit card settings
// are kept. Returns the number of transactions deleted.
func (r *LedgerRepository) DeleteUserTransactionsTx(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
) (int64, error) {

	if _, err := tx.Exec(ctx, `
		DELETE FROM credit_card_statements s
		USING accounts a
		WHERE a.id = s.credit_card_id
		  AND a.user_id = $1
	`, userID); err != nil {
		return 0, err
	}

	// user_id can be NULL on older rows, so also match through the entries' accounts.
	tag, err := tx.Exec(ctx, `
		DELETE FROM transactions t
		WHERE t.user_id = $1
		   OR EXISTS (
				SELECT 1
				FROM entries e
				JOIN accounts a ON a.id = e.account_id
				WHERE e.transaction_id = t.id
				  AND a.user_id = $1
		   )
	`, userID)
	if err != nil {
		return 0, err
	}

	// After the transactions, so imported rows are counted above rather than cascading away here.
	if _, err := tx.Exec(ctx, `
		DELETE FROM bank_statements s
		USING accounts a
		WHERE a.id = s.account_id
		  AND a.user_id = $1
	`, userID); err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
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

func (r *LedgerRepository) GetUserAccounts(
	ctx context.Context,
	userID uuid.UUID,
) ([]Account, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			a.id,
			a.user_id,
			a.name,
			a.type,
			a.created_at,
			COALESCE(
				SUM(
					CASE
						WHEN e.type = 'DEBIT' THEN e.amount
						ELSE -e.amount
					END
				),
				0
			) AS balance
		FROM accounts a
		LEFT JOIN entries e
			ON e.account_id = a.id
		WHERE a.user_id = $1
		GROUP BY
			a.id,
			a.user_id,
			a.name,
			a.type,
			a.created_at
		ORDER BY a.created_at, a.id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account

	for rows.Next() {
		var account Account

		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Name,
			&account.Type,
			&account.CreatedAt,
			&account.Balance,
		)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *LedgerRepository) GetUserRecentTransactions(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]RecentTransaction, error) {

	rows, err := r.db.Query(
		ctx,
		`
		WITH recent_transactions AS (
			SELECT DISTINCT
				t.id,
				t.description,
				t.created_at
			FROM transactions t
			JOIN entries e
				ON e.transaction_id = t.id
			JOIN accounts a
				ON a.id = e.account_id
			WHERE a.user_id = $1
			ORDER BY t.created_at DESC
			LIMIT $2
		)

		SELECT
			rt.id,
			rt.description,
			rt.created_at,

			chosen.amount,

			CASE
				WHEN EXISTS (
					SELECT 1
					FROM entries e2
					JOIN accounts a2
						ON a2.id = e2.account_id
					WHERE e2.transaction_id = rt.id
						AND a2.user_id = $1
						AND a2.type = 'ASSET'
						AND e2.type = 'DEBIT'
				)
				AND EXISTS (
					SELECT 1
					FROM entries e3
					JOIN accounts a3
						ON a3.id = e3.account_id
					WHERE e3.transaction_id = rt.id
						AND a3.user_id = $1
						AND a3.type = 'ASSET'
						AND e3.type = 'CREDIT'
				)
				THEN 'TRANSFER'

				WHEN chosen.account_type = 'ASSET'
					AND chosen.entry_type = 'DEBIT'
					THEN 'IN'

				WHEN chosen.account_type = 'ASSET'
					AND chosen.entry_type = 'CREDIT'
					THEN 'OUT'

				WHEN chosen.account_type IN (
					'LIABILITY',
					'CREDIT_CARD',
					'BNPL'
				)
					AND chosen.entry_type = 'DEBIT'
					THEN 'IN'

				WHEN chosen.account_type IN (
					'LIABILITY',
					'CREDIT_CARD',
					'BNPL'
				)
					AND chosen.entry_type = 'CREDIT'
					THEN 'OUT'

				WHEN chosen.account_type = 'INCOME'
					AND chosen.entry_type = 'CREDIT'
					THEN 'IN'

				WHEN chosen.account_type = 'EXPENSE'
					AND chosen.entry_type = 'DEBIT'
					THEN 'OUT'

				ELSE 'TRANSFER'
			END AS direction,

			chosen.account_name

		FROM recent_transactions rt

		JOIN LATERAL (
			SELECT
				e.amount,
				e.type AS entry_type,
				a.type AS account_type,
				a.name AS account_name
			FROM entries e
			JOIN accounts a
				ON a.id = e.account_id
			WHERE e.transaction_id = rt.id
				AND a.user_id = $1

			ORDER BY
				CASE
					WHEN a.type = 'ASSET' THEN 1
					WHEN a.type IN (
						'CREDIT_CARD',
						'LIABILITY',
						'BNPL'
					) THEN 2
					WHEN a.type IN (
						'INCOME',
						'EXPENSE'
					) THEN 3
					ELSE 4
				END

			LIMIT 1
		) chosen ON TRUE

		ORDER BY rt.created_at DESC
		`,
		userID,
		limit,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var transactions []RecentTransaction

	for rows.Next() {
		var transaction RecentTransaction

		err := rows.Scan(
			&transaction.ID,
			&transaction.Description,
			&transaction.CreatedAt,
			&transaction.Amount,
			&transaction.Direction,
			&transaction.AccountName,
		)

		if err != nil {
			return nil, err
		}

		transactions = append(
			transactions,
			transaction,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *LedgerRepository) GetTransaction(
	ctx context.Context,
	transactionID uuid.UUID,
) (*TransactionDetail, error) {

	var transaction TransactionDetail

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			description,
			created_at
		FROM transactions
		WHERE id = $1
		`,
		transactionID,
	).Scan(
		&transaction.ID,
		&transaction.Description,
		&transaction.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			e.id,
			e.transaction_id,
			e.account_id,
			a.name,
			a.type,
			e.amount,
			e.type
		FROM entries e
		JOIN accounts a
			ON a.id = e.account_id
		WHERE e.transaction_id = $1
		ORDER BY e.id
		`,
		transactionID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {

		var entry TransactionEntry

		err := rows.Scan(
			&entry.ID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.AccountName,
			&entry.AccountType,
			&entry.Amount,
			&entry.Type,
		)

		if err != nil {
			return nil, err
		}

		transaction.Entries = append(
			transaction.Entries,
			entry,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (r *LedgerRepository) GetUserTransactions(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]TransactionDetail, int, error) {

	var total int

	err := r.db.QueryRow(
		ctx,
		`
        SELECT COUNT(*)
        FROM transactions
        WHERE user_id = $1
        `,
		userID,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		ctx,
		`
	WITH paginated_transactions AS (
		SELECT
			t.id,
			t.description,
			t.created_at
		FROM transactions t
		WHERE t.user_id = $1
		ORDER BY
			t.created_at DESC,
			t.id DESC
		LIMIT $2
		OFFSET $3
	)

	SELECT
		pt.id,
		pt.description,
		pt.created_at,

		e.id,
		e.transaction_id,
		e.account_id,
		a.name,
		a.type,
		e.amount,
		e.type

	FROM paginated_transactions pt

	LEFT JOIN entries e
		ON e.transaction_id = pt.id

	LEFT JOIN accounts a
		ON a.id = e.account_id

	ORDER BY
		pt.created_at DESC,
		pt.id DESC,
		e.id

	`,
		userID,
		limit,
		offset,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	transactionMap := make(map[uuid.UUID]int)
	transactions := make([]TransactionDetail, 0)

	for rows.Next() {

		var (
			transactionID uuid.UUID
			description   string
			createdAt     time.Time

			entryID     *uuid.UUID
			entryTxID   *uuid.UUID
			accountID   *uuid.UUID
			accountName *string
			accountType *AccountType
			amount      *decimal.Decimal
			entryType   *EntryType
		)

		err := rows.Scan(
			&transactionID,
			&description,
			&createdAt,

			&entryID,
			&entryTxID,
			&accountID,
			&accountName,
			&accountType,
			&amount,
			&entryType,
		)

		if err != nil {
			return nil, 0, err
		}

		index, exists := transactionMap[transactionID]

		if !exists {

			index = len(transactions)

			transactionMap[transactionID] = index

			transactions = append(
				transactions,
				TransactionDetail{
					ID:          transactionID,
					Description: description,
					CreatedAt:   createdAt,
					Entries:     make([]TransactionEntry, 0),
				},
			)
		}

		if entryID != nil {

			transactions[index].Entries = append(
				transactions[index].Entries,
				TransactionEntry{
					ID:            *entryID,
					TransactionID: *entryTxID,
					AccountID:     *accountID,
					AccountName:   *accountName,
					AccountType:   *accountType,
					Amount:        *amount,
					Type:          *entryType,
				},
			)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// MoneyFlow is money that moved from one account (the credited side) to another (the debited side).
type MoneyFlow struct {
	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	Amount        decimal.Decimal
	Transactions  int
}

// GetMoneyFlows totals the user's money movements per (from, to) account pair, for transactions
// dated in [from, to) when those are given. Opening balances are a starting point, not a movement.
// shortcut: only two-entry transactions count (every import and hand entry); a split transaction
// has no single source for each destination, so it's left out until one exists.
func (r *LedgerRepository) GetMoneyFlows(ctx context.Context, userID uuid.UUID, from, to *time.Time) ([]MoneyFlow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.account_id, d.account_id, SUM(d.amount), COUNT(*)
		FROM transactions t
		JOIN entries d ON d.transaction_id = t.id AND d.type = 'DEBIT'
		JOIN entries c ON c.transaction_id = t.id AND c.type = 'CREDIT'
		WHERE t.user_id = $1
		  AND NOT t.opening_balance
		  AND ($2::timestamptz IS NULL OR t.created_at >= $2)
		  AND ($3::timestamptz IS NULL OR t.created_at < $3)
		  AND (SELECT count(*) FROM entries x WHERE x.transaction_id = t.id) = 2
		GROUP BY 1, 2
		ORDER BY 3 DESC
	`, userID, from, to)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (MoneyFlow, error) {
		var f MoneyFlow
		err := row.Scan(&f.FromAccountID, &f.ToAccountID, &f.Amount, &f.Transactions)
		return f, err
	})
}
