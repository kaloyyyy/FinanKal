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
