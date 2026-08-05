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

func NewStatementRepository(
	db *pgxpool.Pool,
) *StatementRepository {

	return &StatementRepository{
		db: db,
	}
}

// ===============================
// Create
// ===============================

func (r *StatementRepository) CreateStatement(
	ctx context.Context,
	statement CreditCardStatement,
) (uuid.UUID, error) {

	var id uuid.UUID

	err := r.db.QueryRow(
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

// ===============================
// Find
// ===============================

func (r *StatementRepository) FindStatementByCycle(
	ctx context.Context,
	cardID uuid.UUID,
	cycle BillingCycle,
) (*CreditCardStatement, error) {

	var statement CreditCardStatement

	err := r.db.QueryRow(
		ctx,
		`
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
		WHERE credit_card_id=$1
		AND start_date=$2
		AND end_date=$3
		`,
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

func (r *StatementRepository) FindStatementByCycleTx(
	ctx context.Context,
	tx pgx.Tx,
	cardID uuid.UUID,
	cycle BillingCycle,
) (*CreditCardStatement, error) {

	var statement CreditCardStatement

	err := tx.QueryRow(
		ctx,
		`
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
		WHERE credit_card_id=$1
		AND start_date=$2
		AND end_date=$3
		`,
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

// ===============================
// Get
// ===============================

func (r *StatementRepository) GetStatement(
	ctx context.Context,
	statementID uuid.UUID,
) (*CreditCardStatement, error) {

	var statement CreditCardStatement

	err := r.db.QueryRow(
		ctx,
		`
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
		WHERE id=$1
		`,
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

	return &statement, nil
}

func (r *StatementRepository) GetStatementTx(
	ctx context.Context,
	tx pgx.Tx,
	statementID uuid.UUID,
) (*CreditCardStatement, error) {

	var statement CreditCardStatement

	err := tx.QueryRow(
		ctx,
		`
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
		WHERE id=$1
		`,
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

	return &statement, nil
}

// ===============================
// Update
// ===============================

func (r *StatementRepository) UpdateStatementStatus(
	ctx context.Context,
	statementID uuid.UUID,
	status StatementStatus,
) error {

	_, err := r.db.Exec(
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

func (r *StatementRepository) IncrementStatementTotal(
	ctx context.Context,
	statementID uuid.UUID,
	amount decimal.Decimal,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		UPDATE credit_card_statements
		SET total_amount = total_amount + $2,
		    updated_at=NOW()
		WHERE id=$1
		`,
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
		SET total_amount = total_amount + $2,
		    updated_at=NOW()
		WHERE id=$1
		`,
		statementID,
		amount,
	)

	return err
}

func (r *StatementRepository) UpdateStatementTotal(
	ctx context.Context,
	statementID uuid.UUID,
	total decimal.Decimal,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		UPDATE credit_card_statements
		SET
			total_amount = $2,
			updated_at = NOW()
		WHERE id = $1
		`,
		statementID,
		total,
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
		SET
			total_amount = $2,
			updated_at = NOW()
		WHERE id = $1
		`,
		statementID,
		total,
	)

	return err
}

func (r *StatementRepository) GetPaidStatementsToClose(
	ctx context.Context,
) ([]CreditCardStatement, error) {

	rows, err :=
		r.db.Query(
			ctx,
			`
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
			WHERE status = 'PAID'
			AND due_date < NOW()
			`,
		)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var statements []CreditCardStatement

	for rows.Next() {

		var statement CreditCardStatement

		err :=
			rows.Scan(
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

		statements =
			append(statements, statement)
	}

	return statements, rows.Err()
}

func (r *StatementRepository) GetOpenStatement(
	ctx context.Context,
	cardID uuid.UUID,
) (*CreditCardStatement, error) {

	var statement CreditCardStatement

	err :=
		r.db.QueryRow(
			ctx,
			`
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
			WHERE credit_card_id = $1
			AND status = 'OPEN'
			ORDER BY created_at DESC
			LIMIT 1
			`,
			cardID,
		).
			Scan(
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
