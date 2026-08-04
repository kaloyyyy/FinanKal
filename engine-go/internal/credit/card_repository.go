package credit

import (
	"context"
	_ "time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type CardRepository struct {
	db *pgxpool.Pool
}

func NewCardRepository(db *pgxpool.Pool) *CardRepository {
	return &CardRepository{
		db: db,
	}
}

func (r *CardRepository) CreateCreditCard(
	ctx context.Context,
	accountID uuid.UUID,
	creditLimit decimal.Decimal,
	billingDay int,
	dueDay int,
) (uuid.UUID, error) {

	var cardID uuid.UUID

	err := r.db.QueryRow(ctx, `
		INSERT INTO credit_cards (account_id, credit_limit, billing_day, payment_due_days)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`,
		accountID, creditLimit, billingDay, dueDay,
	).Scan(&cardID)

	return cardID, err
}

func (r *CardRepository) GetCreditCard(
	ctx context.Context,
	cardID uuid.UUID,
) (*CreditCard, error) {

	var card CreditCard

	err := r.db.QueryRow(ctx, `
		SELECT id, account_id, credit_limit, billing_day, payment_due_days, cutoff_time, created_at, 
			updated_at 
		FROM credit_cards WHERE id = $1
	`, cardID).Scan(
		&card.ID,
		&card.AccountID,
		&card.CreditLimit,
		&card.BillingDay,
		&card.PaymentDueDays,
		&card.CutoffTime,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (r *CardRepository) ListCreditCards(
	ctx context.Context,
	userID uuid.UUID,
) ([]CreditCard, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			cc.id,
			cc.account_id,
			cc.credit_limit,
			cc.billing_day,
			cc.payment_due_days,
			cc.cutoff_time,
			cc.created_at,
			cc.updated_at
		FROM credit_cards cc
		JOIN accounts a
			ON cc.account_id = a.id
		WHERE a.user_id = $1
		ORDER BY cc.created_at
	`, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cards []CreditCard

	for rows.Next() {

		var card CreditCard

		err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.CreditLimit,
			&card.BillingDay,
			&card.PaymentDueDays,
			&card.CutoffTime,
			&card.CreatedAt,
			&card.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}

	return cards, rows.Err()
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

	_, err := r.db.Exec(ctx,
		`DELETE FROM credit_cards WHERE id = $1`,
		cardID,
	)

	return err
}

func (r *CardRepository) GetCreditCardByAccount(
	ctx context.Context,
	accountID uuid.UUID,
) (*CreditCard, error) {

	var card CreditCard

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			account_id,
			credit_limit,
			billing_day,
			payment_due_days,
			cutoff_time,
			created_at,
			updated_at
		FROM credit_cards
		WHERE account_id = $1
	`, accountID).Scan(
		&card.ID,
		&card.AccountID,
		&card.CreditLimit,
		&card.BillingDay,
		&card.PaymentDueDays,
		&card.CutoffTime,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (r *CardRepository) GetAccountType(
	ctx context.Context,
	accountID uuid.UUID,
) (string, error) {

	var accountType string

	err := r.db.QueryRow(ctx,
		`SELECT type FROM accounts WHERE id = $1`,
		accountID,
	).Scan(&accountType)

	return accountType, err
}

func (r *CardRepository) GetUserCreditCards(
	ctx context.Context,
	userID uuid.UUID,
) ([]CreditCard, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			cc.id,
			cc.account_id,
			cc.credit_limit,
			cc.billing_day,
			cc.payment_due_days,
			cc.cutoff_time,
			cc.created_at,
			cc.updated_at
		FROM credit_cards cc
		JOIN accounts a
			ON cc.account_id = a.id
		WHERE a.user_id = $1
		ORDER BY cc.created_at DESC
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cards []CreditCard

	for rows.Next() {

		var card CreditCard

		err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.CreditLimit,
			&card.BillingDay,
			&card.PaymentDueDays,
			&card.CutoffTime,
			&card.CreatedAt,
			&card.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}

	return cards, rows.Err()
}
