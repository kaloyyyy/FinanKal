package credit

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

const cardColumns = `
	cc.id,
	cc.clearing_account_id,
	cc.credit_limit,
	cc.billing_day,
	cc.payment_due_days,
	cc.created_at,
	cc.updated_at
`

type CardRepository struct {
	db *pgxpool.Pool
}

func NewCardRepository(db *pgxpool.Pool) *CardRepository {
	return &CardRepository{db: db}
}

func scanCard(row pgx.Row) (CreditCard, error) {
	var card CreditCard

	err := row.Scan(
		&card.CardID,
		&card.ClearingAccountId,
		&card.CreditLimit,
		&card.BillingDay,
		&card.PaymentDueDays,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	return card, err
}

func (r *CardRepository) CreateCreditCardTx(
	ctx context.Context,
	tx pgx.Tx,
	accountID uuid.UUID,
	clearingAccountID uuid.UUID,
	creditLimit decimal.Decimal,
	billingDay int,
	dueDay int,
) (uuid.UUID, error) {

	var cardID uuid.UUID

	err := tx.QueryRow(ctx, `
		INSERT INTO credit_cards (id, clearing_account_id, credit_limit, billing_day, payment_due_days)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`,
		accountID, clearingAccountID, creditLimit, billingDay, dueDay,
	).Scan(&cardID)

	return cardID, err
}

func (r *CardRepository) GetCreditCard(
	ctx context.Context,
	cardID uuid.UUID,
) (*CreditCard, error) {

	card, err := scanCard(r.db.QueryRow(ctx, `SELECT `+cardColumns+` FROM credit_cards cc WHERE cc.id = $1`, cardID))
	if err != nil {
		return nil, err
	}

	return &card, nil
}

// LockCreditLimitTx locks the card row until tx ends and returns its current limit.
func (r *CardRepository) LockCreditLimitTx(
	ctx context.Context,
	tx pgx.Tx,
	cardID uuid.UUID,
) (decimal.Decimal, error) {

	var limit decimal.Decimal

	err := tx.QueryRow(ctx, `SELECT credit_limit FROM credit_cards WHERE id = $1 FOR UPDATE`, cardID).Scan(&limit)

	return limit, err
}

func (r *CardRepository) ListCreditCards(
	ctx context.Context,
	userID uuid.UUID,
) ([]CreditCard, error) {

	rows, err := r.db.Query(ctx, `
		SELECT `+cardColumns+`
		FROM credit_cards cc
		JOIN accounts a
			ON cc.id = a.id
		WHERE a.user_id = $1
		ORDER BY cc.created_at
	`, userID)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (CreditCard, error) {
		return scanCard(row)
	})
}

func (r *CardRepository) UpdateCreditCard(
	ctx context.Context,
	cardID uuid.UUID,
	creditLimit decimal.Decimal,
	billingDay int,
	paymentDueDays int,
) error {

	_, err := r.db.Exec(ctx, `
		UPDATE credit_cards
		SET
			credit_limit = $2,
			billing_day = $3,
			payment_due_days = $4,
			updated_at = NOW()
		WHERE id = $1
	`,
		cardID,
		creditLimit,
		billingDay,
		paymentDueDays,
	)

	return err
}

func (r *CardRepository) DeleteCreditCard(
	ctx context.Context,
	cardID uuid.UUID,
) error {

	_, err := r.db.Exec(ctx, `DELETE FROM credit_cards WHERE id = $1`, cardID)

	return err
}
