package credit

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type StatementRepository struct {
	db *pgxpool.Pool
}

func NewStatementRepository(db *pgxpool.Pool) *StatementRepository {
	return &StatementRepository{
		db: db,
	}
}

func (r *StatementRepository) CreateStatement(
	ctx context.Context,
	statement CreditCardStatement,
) (uuid.UUID, error) {

	var id uuid.UUID

	const query = `
INSERT INTO credit_card_statements (
	credit_card_id,
	start_date,
	end_date,
	statement_date,
	due_date,
	total_amount,
	status
)
VALUES (
	$1, $2, $3, $4, $5, $6, $7
)
RETURNING id
`

	err := r.db.QueryRow(
		ctx,
		query,
		statement.CreditCardID,
		statement.StartDate,
		statement.EndDate,
		statement.StatementDate,
		statement.DueDate,
		statement.TotalAmount,
		statement.Status,
	).Scan(&id)

	return id, err
}

func (r *StatementRepository) FindStatementByCycle(
	ctx context.Context,
	cardID uuid.UUID,
	cycle BillingCycle,
) (*CreditCardStatement, error) {

	var statement CreditCardStatement

	const query = `
SELECT
	id,
	credit_card_id,
	start_date,
	end_date,
	statement_date,
	due_date,
	total_amount,
	status,
	created_at,
	updated_at
FROM credit_card_statements
WHERE
	credit_card_id = $1
	AND start_date = $2
	AND end_date = $3
`

	err := r.db.QueryRow(
		ctx,
		query,
		cardID,
		cycle.CycleStartDate,
		cycle.CycleEndDate,
	).Scan(
		&statement.ID,
		&statement.CreditCardID,
		&statement.StartDate,
		&statement.EndDate,
		&statement.StatementDate,
		&statement.DueDate,
		&statement.TotalAmount,
		&statement.Status,
		&statement.CreatedAt,
		&statement.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &statement, nil
}

func (r *StatementRepository) UpdateStatementTotal(
	ctx context.Context,
	statementID uuid.UUID,
	total decimal.Decimal,
) error {

	const query = `
UPDATE credit_card_statements
SET
	total_amount = $2,
	updated_at = NOW()
WHERE id = $1
`

	_, err := r.db.Exec(
		ctx,
		query,
		statementID,
		total,
	)

	return err
}

func (r *StatementRepository) UpdateStatementStatus(
	ctx context.Context,
	statementID uuid.UUID,
	status StatementStatus,
) error {

	const query = `
UPDATE credit_card_statements
SET
	status = $2,
	updated_at = NOW()
WHERE id = $1
`

	_, err := r.db.Exec(
		ctx,
		query,
		statementID,
		status,
	)

	return err
}

func (r *StatementRepository) GetStatement(
	ctx context.Context,
	statementID uuid.UUID,
) (*CreditCardStatement, error) {

	var statement CreditCardStatement

	const query = `
		SELECT
			id,
			credit_card_id,
			start_date,
			end_date,
			statement_date,
			due_date,
			total_amount,
			status,
			created_at,
			updated_at
		FROM credit_card_statements
		WHERE id = $1
		`

	err := r.db.QueryRow(
		ctx,
		query,
		statementID,
	).Scan(
		&statement.ID,
		&statement.CreditCardID,
		&statement.StartDate,
		&statement.EndDate,
		&statement.StatementDate,
		&statement.DueDate,
		&statement.TotalAmount,
		&statement.Status,
		&statement.CreatedAt,
		&statement.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &statement, err
}
func (r *StatementRepository) IncrementStatementTotal(
	ctx context.Context,
	statementID uuid.UUID,
	amount decimal.Decimal,
) error {

	const query = `
		UPDATE credit_card_statements
		SET
			total_amount = total_amount + $2,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		statementID,
		amount,
	)

	return err
}

func (r *StatementRepository) IncrementStatementTotalTx(
	ctx context.Context,
	tx pgx.Tx,
	statementID uuid.UUID,
	amount decimal.Decimal,
) error {

	_, err := tx.Exec(
		ctx,
		`
		UPDATE credit_card_statements
		SET
			total_amount = total_amount + $2,
			updated_at = NOW()
		WHERE id = $1
		`,
		statementID,
		amount,
	)

	return err
}
