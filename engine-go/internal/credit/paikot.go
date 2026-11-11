package credit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// PaikotLink says dues on an account can be paid with a card (PayDirect, QRph), at a fee.
type PaikotLink struct {
	Card       CreditCard
	CardName   string
	FeePercent decimal.Decimal
}

// Paikot is a due moved onto a card: charge the card from PayFrom to PayBy and it's billed on the
// card's next statement, due DueDate.
type Paikot struct {
	PayFrom, PayBy, DueDate time.Time
	Fee                     decimal.Decimal
}

// Roll works out where a due lands when the card pays it. The charge belongs to the card's cycle the
// due date falls in (today's, when overdue): made after the previous statement date, so it's billed
// on the next one.
// shortcut: one hop and no credit-limit check; walk the links and the limit once chains (A pays B pays C) need forecasting.
func (l PaikotLink) Roll(due, today time.Time, amount decimal.Decimal) Paikot {
	payBy := due
	if payBy.Before(today) {
		payBy = today
	}
	cycle := CalculateBillingCycle(payBy, l.Card)
	payFrom := cycle.CycleStartDate
	if payFrom.Before(today) {
		payFrom = today
	}
	return Paikot{
		PayFrom: payFrom,
		PayBy:   payBy,
		DueDate: cycle.DueDate,
		Fee:     amount.Mul(l.FeePercent).Div(decimal.NewFromInt(100)).Round(2),
	}
}

// PaikotLinks returns the user's links by the account whose dues they move. Cards without a billing
// day have no cycle to move a due into, so their links are left out.
func (s *Service) PaikotLinks(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]PaikotLink, error) {
	rows, err := s.statementRepo.db.Query(ctx, `
		SELECT p.account_id, p.fee_percent, a.name, cc.id, cc.billing_day, cc.payment_due_days, cc.due_rule, cc.statement_shift
		FROM paikot_links p
		JOIN credit_cards cc ON cc.id = p.card_id
		JOIN accounts a ON a.id = cc.id
		WHERE a.user_id = $1 AND cc.billing_day > 0
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("paikot links: %w", err)
	}
	links := map[uuid.UUID]PaikotLink{}
	var accountID uuid.UUID
	var l PaikotLink
	_, err = pgx.ForEachRow(rows, []any{&accountID, &l.FeePercent, &l.CardName, &l.Card.CardID, &l.Card.BillingDay,
		&l.Card.PaymentDueDays, &l.Card.DueRule, &l.Card.StatementShift}, func() error {
		links[accountID] = l
		return nil
	})
	return links, err
}

// SetPaikotLink lets cardID pay the dues of accountID (one of the user's loans or other cards), replacing
// the account's earlier link. uuid.Nil as cardID removes it.
func (s *Service) SetPaikotLink(ctx context.Context, userID, accountID, cardID uuid.UUID, feePercent decimal.Decimal) error {
	db := s.statementRepo.db
	if cardID == uuid.Nil {
		_, err := db.Exec(ctx, `
			DELETE FROM paikot_links p USING accounts a
			WHERE a.id = p.account_id AND a.user_id = $1 AND p.account_id = $2
		`, userID, accountID)
		return err
	}
	if feePercent.IsNegative() || feePercent.GreaterThan(decimal.NewFromInt(100)) {
		return fmt.Errorf("%w: fee must be between 0 and 100 percent", ErrInvalidRequest)
	}
	res, err := db.Exec(ctx, `
		INSERT INTO paikot_links (account_id, card_id, fee_percent)
		SELECT d.id, cc.id, $4
		FROM accounts d
		CROSS JOIN credit_cards cc
		JOIN accounts c ON c.id = cc.id
		WHERE d.id = $2 AND cc.id = $3 AND d.id <> cc.id
		  AND d.user_id = $1 AND c.user_id = $1
		  AND d.type IN ('LIABILITY', 'CREDIT_CARD')
		ON CONFLICT (account_id) DO UPDATE SET card_id = EXCLUDED.card_id, fee_percent = EXCLUDED.fee_percent
	`, userID, accountID, cardID, feePercent)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%w: link one of your loans or cards to a different card of yours", ErrInvalidRequest)
	}
	return nil
}
