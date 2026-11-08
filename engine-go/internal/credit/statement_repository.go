package credit

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

const selectStatement = `
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
`

type StatementRepository struct {
	db *pgxpool.Pool
}

func NewStatementRepository(db *pgxpool.Pool) *StatementRepository {
	return &StatementRepository{db: db}
}

func scanStatement(row pgx.Row) (*CreditCardStatement, error) {
	var statement CreditCardStatement

	err := row.Scan(
		&statement.ID,
		&statement.CardID,
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

func (r *StatementRepository) CreateStatementTx(
	ctx context.Context,
	tx pgx.Tx,
	statement CreditCardStatement,
) (uuid.UUID, error) {

	var id uuid.UUID

	err := tx.QueryRow(
		ctx,
		`
		INSERT INTO credit_card_statements (
			credit_card_id,
			start_date,
			end_date,
			statement_date,
			due_date,
			total_amount,
			status
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
		`,
		statement.CardID,
		statement.StartDate,
		statement.EndDate,
		statement.StatementDate,
		statement.DueDate,
		statement.TotalAmount,
		statement.Status,
	).Scan(&id)

	return id, err
}

func (r *StatementRepository) FindStatementByCycleTx(
	ctx context.Context,
	tx pgx.Tx,
	cardID uuid.UUID,
	cycle BillingCycle,
) (*CreditCardStatement, error) {

	return scanStatement(tx.QueryRow(
		ctx,
		selectStatement+`WHERE credit_card_id=$1 AND start_date=$2 AND end_date=$3`,
		cardID,
		cycle.CycleStartDate,
		cycle.CycleEndDate,
	))
}

func (r *StatementRepository) GetStatement(
	ctx context.Context,
	statementID uuid.UUID,
) (*CreditCardStatement, error) {

	return scanStatement(r.db.QueryRow(ctx, selectStatement+`WHERE id=$1`, statementID))
}

func (r *StatementRepository) GetStatementForUpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	statementID uuid.UUID,
) (*CreditCardStatement, error) {

	return scanStatement(tx.QueryRow(ctx, selectStatement+`WHERE id=$1 FOR UPDATE`, statementID))
}

func (r *StatementRepository) UpdateStatementStatusTx(
	ctx context.Context,
	tx pgx.Tx,
	statementID uuid.UUID,
	status StatementStatus,
) error {

	_, err := tx.Exec(
		ctx,
		`
		UPDATE credit_card_statements
		SET status=$2,
		    updated_at=NOW()
		WHERE id=$1
		`,
		statementID,
		status,
	)

	return err
}

// AddPurchaseToStatementTx adds to the amount owed. A PAID statement keeps its paid
// total, so a late purchase reopens it owing only that purchase; otherwise the debt
// could never be paid and the amount already paid would be owed again.
func (r *StatementRepository) AddPurchaseToStatementTx(
	ctx context.Context,
	tx pgx.Tx,
	statementID uuid.UUID,
	amount decimal.Decimal,
) error {

	_, err := tx.Exec(
		ctx,
		`
		UPDATE credit_card_statements
		SET total_amount = CASE WHEN status = 'PAID' THEN $2 ELSE total_amount + $2 END,
		    status = CASE WHEN status = 'PAID' THEN 'OPEN' ELSE status END,
		    updated_at = NOW()
		WHERE id = $1
		`,
		statementID,
		amount,
	)

	return err
}

func (r *StatementRepository) UpdateStatementTotalTx(
	ctx context.Context,
	tx pgx.Tx,
	statementID uuid.UUID,
	total decimal.Decimal,
) error {

	_, err := tx.Exec(
		ctx,
		`
		UPDATE credit_card_statements
		SET total_amount = $2,
		    updated_at = NOW()
		WHERE id = $1
		`,
		statementID,
		total,
	)

	return err
}
