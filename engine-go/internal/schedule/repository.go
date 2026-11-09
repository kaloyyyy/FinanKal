package schedule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

const scheduleColumns = `
	id,
	user_id,
	name,
	schedule_type,
	transaction_type,
	amount,
	source_account_id,
	destination_account_id,
	category_account_id,
	start_date,
	end_date,
	frequency,
	day_of_month,
	active,
	created_at,
	updated_at,
	amount_type
`

const occurrenceColumns = `
	id,
	scheduled_transaction_id,
	transaction_id,
	scheduled_date,
	installment_number,
	amount,
	status,
	created_at
`

type ScheduleRepository struct {
	db *pgxpool.Pool
}

func NewScheduleRepository(
	db *pgxpool.Pool,
) *ScheduleRepository {
	return &ScheduleRepository{
		db: db,
	}
}

func scanSchedule(row pgx.Row) (ScheduledTransaction, error) {
	var s ScheduledTransaction

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.ScheduleType,
		&s.TransactionType,
		&s.Amount,
		&s.SourceAccountID,
		&s.DestinationAccountID,
		&s.CategoryAccountID,
		&s.StartDate,
		&s.EndDate,
		&s.Frequency,
		&s.DayOfMonth,
		&s.Active,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.AmountType,
	)

	return s, err
}

func scanOccurrence(row pgx.Row) (ScheduledTransactionOccurrence, error) {
	var occurrence ScheduledTransactionOccurrence

	err := row.Scan(
		&occurrence.ID,
		&occurrence.ScheduledTransactionID,
		&occurrence.TransactionID,
		&occurrence.ScheduledDate,
		&occurrence.InstallmentNumber,
		&occurrence.Amount,
		&occurrence.Status,
		&occurrence.CreatedAt,
	)

	return occurrence, err
}

func collectOccurrences(rows pgx.Rows, err error) ([]ScheduledTransactionOccurrence, error) {
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (ScheduledTransactionOccurrence, error) {
		return scanOccurrence(row)
	})
}

func (r *ScheduleRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*ScheduledTransaction, error) {

	s, err := scanSchedule(r.db.QueryRow(ctx, `SELECT `+scheduleColumns+` FROM scheduled_transactions WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("scheduled transaction %s not found: %w", id, err)
		}

		return nil, fmt.Errorf("get scheduled transaction: %w", err)
	}

	return &s, nil
}

func (r *ScheduleRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]ScheduledTransaction, error) {

	rows, err := r.db.Query(
		ctx,
		`SELECT `+scheduleColumns+`
		FROM scheduled_transactions
		WHERE user_id = $1
		ORDER BY start_date ASC, created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get scheduled transactions: %w", err)
	}

	results, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ScheduledTransaction, error) {
		return scanSchedule(row)
	})
	if err != nil {
		return nil, fmt.Errorf("scan scheduled transactions: %w", err)
	}

	return results, nil
}

func (r *ScheduleRepository) CreateOccurrence(
	ctx context.Context,
	scheduledTransactionID uuid.UUID,
	scheduledDate time.Time,
	amount *decimal.Decimal,
	installmentNumber *int,
) (*ScheduledTransactionOccurrence, error) {
	query := `
		INSERT INTO scheduled_transaction_occurrences (
			scheduled_transaction_id,
			scheduled_date,
			installment_number,
			amount,
			status
		)
		VALUES ($1, $2, $3, $4, 'PENDING')
		ON CONFLICT (
			scheduled_transaction_id,
			scheduled_date
		)
		DO NOTHING
		RETURNING ` + occurrenceColumns

	occurrence, err := scanOccurrence(r.db.QueryRow(
		ctx,
		query,
		scheduledTransactionID,
		scheduledDate,
		installmentNumber,
		amount,
	))

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf(
			"create scheduled occurrence: %w",
			err,
		)
	}

	return &occurrence, nil
}

func (r *ScheduleRepository) GetOccurrenceByID(
	ctx context.Context,
	id uuid.UUID,
) (*ScheduledTransactionOccurrence, error) {

	occurrence, err := scanOccurrence(r.db.QueryRow(
		ctx,
		`SELECT `+occurrenceColumns+` FROM scheduled_transaction_occurrences WHERE id = $1`,
		id,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("scheduled occurrence %s not found: %w", id, err)
		}

		return nil, fmt.Errorf("get scheduled occurrence: %w", err)
	}

	return &occurrence, nil
}

func (r *ScheduleRepository) GetOccurrences(
	ctx context.Context,
	scheduledTransactionID uuid.UUID,
) ([]ScheduledTransactionOccurrence, error) {

	rows, err := r.db.Query(
		ctx,
		`SELECT `+occurrenceColumns+`
		FROM scheduled_transaction_occurrences
		WHERE scheduled_transaction_id = $1
		ORDER BY scheduled_date ASC`,
		scheduledTransactionID,
	)

	results, err := collectOccurrences(rows, err)
	if err != nil {
		return nil, fmt.Errorf("get scheduled occurrences: %w", err)
	}

	return results, nil
}

func (r *ScheduleRepository) GetPendingOccurrences(
	ctx context.Context,
	userID uuid.UUID,
) ([]ScheduledTransactionOccurrence, error) {

	rows, err := r.db.Query(
		ctx,
		`SELECT
			o.id,
			o.scheduled_transaction_id,
			o.transaction_id,
			o.scheduled_date,
			o.installment_number,
			o.amount,
			o.status,
			o.created_at
		FROM scheduled_transaction_occurrences o
		INNER JOIN scheduled_transactions s
			ON s.id = o.scheduled_transaction_id
		WHERE s.user_id = $1
		  AND o.status = 'PENDING'
		ORDER BY o.scheduled_date ASC`,
		userID,
	)

	results, err := collectOccurrences(rows, err)
	if err != nil {
		return nil, fmt.Errorf("get pending occurrences: %w", err)
	}

	return results, nil
}

func (r *ScheduleRepository) MarkOccurrenceCompleted(
	ctx context.Context,
	occurrenceID uuid.UUID,
	transactionID uuid.UUID,
) error {
	const query = `
		UPDATE scheduled_transaction_occurrences
		SET
			transaction_id = $2,
			status = 'COMPLETED'
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		occurrenceID,
		transactionID,
	)
	if err != nil {
		return fmt.Errorf("complete scheduled occurrence: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf(
			"scheduled occurrence %s not found",
			occurrenceID,
		)
	}

	return nil
}

func (r *ScheduleRepository) MarkOccurrenceFailed(
	ctx context.Context,
	occurrenceID uuid.UUID,
) error {
	const query = `
		UPDATE scheduled_transaction_occurrences
		SET status = 'FAILED'
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, occurrenceID)
	if err != nil {
		return fmt.Errorf("fail scheduled occurrence: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf(
			"scheduled occurrence %s not found",
			occurrenceID,
		)
	}

	return nil
}

// CreateInstallments inserts a monthly TRANSFER schedule from sourceID into accountID with one pending
// occurrence per installment, numbered from 1, all or nothing.
func (r *ScheduleRepository) CreateInstallments(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	sourceID, accountID uuid.UUID,
	installments []Installment,
) (uuid.UUID, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin create installments: %w", err)
	}
	defer tx.Rollback(ctx)

	// Both accounts must be the user's, and the loan must be owed money, or GetPayables won't list it.
	var matched int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM accounts
		WHERE user_id = $1
		  AND ((id = $2 AND type IN ('LIABILITY', 'CREDIT_CARD')) OR (id = $3 AND type = 'ASSET'))
	`, userID, accountID, sourceID).Scan(&matched); err != nil {
		return uuid.Nil, fmt.Errorf("check installment accounts: %w", err)
	}
	if matched != 2 {
		return uuid.Nil, fmt.Errorf("%w: pay from one of your bank or wallet accounts into one of your loans or cards", ErrInvalidSchedule)
	}

	first, last := installments[0], installments[len(installments)-1]
	var id uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO scheduled_transactions (
			user_id, name, schedule_type, transaction_type, amount, source_account_id, destination_account_id,
			start_date, end_date, frequency, day_of_month, amount_type
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`, userID, name, ScheduleTypeRecurring, TransactionTypeTransfer, first.Amount, sourceID, accountID,
		first.DueDate, last.DueDate, FrequencyMonthly, first.DueDate.Day(), AmountTypeFixed,
	).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("create installment schedule: %w", err)
	}

	for i, in := range installments {
		if _, err := tx.Exec(ctx, `
			INSERT INTO scheduled_transaction_occurrences (scheduled_transaction_id, scheduled_date, installment_number, amount, status)
			VALUES ($1, $2, $3, $4, $5)
		`, id, in.DueDate, i+1, in.Amount, OccurrencePending); err != nil {
			return uuid.Nil, fmt.Errorf("create installment %d: %w", i+1, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit installments: %w", err)
	}
	return id, nil
}

// Payable is one unpaid installment of something the user owes (a loan or a card).
type Payable struct {
	Name              string // the schedule, e.g. "Mari Loan ₱33,000 (12 months, …)"
	AccountID         uuid.UUID
	AccountName       string // the loan / card account it pays down
	DueDate           time.Time
	Amount            decimal.Decimal
	InstallmentNumber *int
	Installments      *int // the last installment number recorded, i.e. the loan's term
}

// GetPayables returns the user's unpaid installments, soonest first: pending occurrences of active
// schedules that pay into a liability or credit card account.
func (r *ScheduleRepository) GetPayables(ctx context.Context, userID uuid.UUID) ([]Payable, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.name, a.id, a.name, o.scheduled_date, COALESCE(o.amount, s.amount), o.installment_number,
		       MAX(o.installment_number) OVER (PARTITION BY s.id)
		FROM scheduled_transaction_occurrences o
		JOIN scheduled_transactions s ON s.id = o.scheduled_transaction_id
		JOIN accounts a ON a.id = s.destination_account_id
		WHERE s.user_id = $1
		  AND s.active
		  AND o.status = 'PENDING'
		  AND a.type IN ('LIABILITY', 'CREDIT_CARD')
		ORDER BY o.scheduled_date, s.name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get payables: %w", err)
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Payable, error) {
		var p Payable
		err := row.Scan(&p.Name, &p.AccountID, &p.AccountName, &p.DueDate, &p.Amount, &p.InstallmentNumber, &p.Installments)
		return p, err
	})
}
