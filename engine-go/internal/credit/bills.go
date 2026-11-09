package credit

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Bill is one billing cycle of a card (a credit_card_statements row) and what's still owed on it.
// Whether it's paid isn't stored: it's worked out from the payments that reached the card, so it's
// right whatever order statements are imported in, and after any correction to the ledger.
type Bill struct {
	CardID        uuid.UUID
	CardName      string
	StartDate     time.Time
	StatementDate time.Time
	DueDate       time.Time // printed on the statement, or from the card's rule
	Billed        decimal.Decimal
	Owed          decimal.Decimal
	SoFar         bool // the current cycle, not billed yet: its amount grows until the statement date
}

// CardPayment is money that reached a card: a payment, a joined transfer, a refund.
type CardPayment struct {
	CardID uuid.UUID
	Date   time.Time // calendar date
	Amount decimal.Decimal
}

// AllocatePayments settles each card's bills oldest first. A payment counts toward bills issued before
// it (statement date strictly earlier; a payment on the statement date is already on that statement,
// paying the one before). What's left after a bill is covered carries on to the next. Payments made
// before a card's first tracked bill paid older balances, so they're ignored.
func AllocatePayments(bills []Bill, payments []CardPayment) []Bill {
	out := make([]Bill, len(bills))
	copy(out, bills)
	sort.SliceStable(out, func(i, j int) bool { return out[i].StatementDate.Before(out[j].StatementDate) })
	for i := range out {
		out[i].Owed = out[i].Billed
	}

	sorted := make([]CardPayment, len(payments))
	copy(sorted, payments)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Date.Before(sorted[j].Date) })

	for _, p := range sorted {
		left := p.Amount
		for i := range out {
			b := &out[i]
			if b.CardID != p.CardID || !b.StatementDate.Before(p.Date) || !b.Owed.IsPositive() {
				continue
			}
			paid := decimal.Min(left, b.Owed)
			b.Owed = b.Owed.Sub(paid)
			left = left.Sub(paid)
			if !left.IsPositive() {
				break
			}
		}
	}
	return out
}

// OpenBills returns the user's card bills that still have something owed, soonest due first. Each card
// also gets its current cycle "so far": whatever it owes beyond its open bills (purchases not on a
// statement yet), due by the card's rule after its next statement date. That's next month's bill.
func (s *Service) OpenBills(ctx context.Context, userID uuid.UUID, today time.Time) ([]Bill, error) {
	bills, err := s.statementRepo.userBills(ctx, userID)
	if err != nil {
		return nil, err
	}
	payments, err := s.statementRepo.userCardPayments(ctx, userID)
	if err != nil {
		return nil, err
	}
	cards, err := s.statementRepo.userCards(ctx, userID)
	if err != nil {
		return nil, err
	}

	var open []Bill
	for _, b := range AllocatePayments(bills, payments) {
		if b.Owed.IsPositive() {
			open = append(open, b)
		}
	}
	cycles := CurrentCycles(cards, open, bills, today)
	// Next month's bill is kept on record as an estimate until its statement replaces it.
	if err := s.statementRepo.saveEstimates(ctx, userID, cycles); err != nil {
		return nil, err
	}
	open = append(open, cycles...)
	sort.SliceStable(open, func(i, j int) bool { return open[i].DueDate.Before(open[j].DueDate) })
	return open, nil
}

// saveEstimates records each card's current cycle as an ESTIMATED bill, replacing the user's earlier
// estimates, so next month's bill is on record (with its statement and due dates) before the statement
// arrives. Estimates are worked out again on every read, so they follow new purchases and payments.
// shortcut: rewrites the estimates on each payables read; fine at a handful of cards, move to the import path if reads get hot.
func (r *StatementRepository) saveEstimates(ctx context.Context, userID uuid.UUID, cycles []Bill) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		DELETE FROM credit_card_statements s
		USING accounts a
		WHERE a.id = s.credit_card_id AND a.user_id = $1 AND s.status = $2
	`, userID, StatementEstimated); err != nil {
		return fmt.Errorf("clear estimates: %w", err)
	}
	for _, c := range cycles {
		// A real bill on the same statement date wins: the estimate is only for cycles not billed yet.
		if _, err := tx.Exec(ctx, `
			INSERT INTO credit_card_statements (credit_card_id, start_date, end_date, statement_date, due_date, total_amount, status)
			VALUES ($1, $2, $3, $3, $4, $5, $6)
			ON CONFLICT (credit_card_id, statement_date) DO NOTHING
		`, c.CardID, c.StartDate, c.StatementDate, c.DueDate, c.Owed, StatementEstimated); err != nil {
			return fmt.Errorf("save estimate: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// CardOwing is a card with its settings and what it owes now (positive).
type CardOwing struct {
	Card CreditCard
	Name string
	Owed decimal.Decimal
}

// CurrentCycles works out each card's bill so far: owed now, minus what its open bills already cover.
// Its statement date is the card's next one after both its latest bill and today.
func CurrentCycles(cards []CardOwing, open, all []Bill, today time.Time) []Bill {
	var cycles []Bill
	for _, c := range cards {
		unbilled := c.Owed
		var latest time.Time
		for _, b := range open {
			if b.CardID == c.Card.CardID {
				unbilled = unbilled.Sub(b.Owed)
			}
		}
		for _, b := range all {
			if b.CardID == c.Card.CardID && b.StatementDate.After(latest) {
				latest = b.StatementDate
			}
		}
		if !unbilled.IsPositive() || c.Card.BillingDay == 0 {
			continue
		}
		// Today's cycle: this month's statement if it hasn't happened yet, else next month's (moved off a
		// weekend by the card's shift). Never one already billed.
		after := today
		if latest.After(after) {
			after = latest
		}
		statement := NextStatementDate(after, c.Card)
		cycles = append(cycles, Bill{
			CardID:        c.Card.CardID,
			CardName:      c.Name,
			StartDate:     CalculateStartDate(statement, c.Card),
			StatementDate: statement,
			DueDate:       CalculateDueDate(statement, c.Card),
			Billed:        unbilled,
			Owed:          unbilled,
			SoFar:         true,
		})
	}
	return cycles
}

// userCards reads the user's credit cards (with settings) and what each owes now.
func (r *StatementRepository) userCards(ctx context.Context, userID uuid.UUID) ([]CardOwing, error) {
	rows, err := r.db.Query(ctx, `
		SELECT cc.id, a.name, cc.billing_day, cc.payment_due_days, cc.due_rule, cc.statement_shift,
		       -COALESCE(SUM(CASE WHEN e.type = 'DEBIT' THEN e.amount ELSE -e.amount END), 0)
		FROM credit_cards cc
		JOIN accounts a ON a.id = cc.id
		LEFT JOIN entries e ON e.account_id = cc.id
		WHERE a.user_id = $1 AND a.type = 'CREDIT_CARD'
		GROUP BY cc.id, a.name, cc.billing_day, cc.payment_due_days, cc.due_rule, cc.statement_shift
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("user cards: %w", err)
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (CardOwing, error) {
		var c CardOwing
		err := row.Scan(&c.Card.CardID, &c.Name, &c.Card.BillingDay, &c.Card.PaymentDueDays, &c.Card.DueRule, &c.Card.StatementShift, &c.Owed)
		return c, err
	})
}

// userBills reads every bill of the user's cards, with the due date filled in from the card's rule
// when the statement didn't print one.
func (r *StatementRepository) userBills(ctx context.Context, userID uuid.UUID) ([]Bill, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.credit_card_id, a.name, s.start_date, s.statement_date, s.due_date, s.total_amount,
		       cc.payment_due_days, cc.due_rule
		FROM credit_card_statements s
		JOIN credit_cards cc ON cc.id = s.credit_card_id
		JOIN accounts a ON a.id = cc.id
		WHERE a.user_id = $1 AND s.status <> 'ESTIMATED' -- estimates are worked out from these, not billed
		ORDER BY s.statement_date
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("user bills: %w", err)
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Bill, error) {
		var b Bill
		var due *time.Time
		var card CreditCard
		err := row.Scan(&b.CardID, &b.CardName, &b.StartDate, &b.StatementDate, &due, &b.Billed, &card.PaymentDueDays, &card.DueRule)
		if due != nil {
			b.DueDate = *due
		} else {
			b.DueDate = CalculateDueDate(b.StatementDate, card)
		}
		return b, err
	})
}

// userCardPayments reads every amount that reached the user's cards (DEBIT entries), by calendar date.
// Opening balances are where the ledger starts, not payments.
func (r *StatementRepository) userCardPayments(ctx context.Context, userID uuid.UUID) ([]CardPayment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT e.account_id, (t.created_at AT TIME ZONE 'Asia/Manila')::date, e.amount
		FROM entries e
		JOIN transactions t ON t.id = e.transaction_id
		JOIN credit_cards cc ON cc.id = e.account_id
		JOIN accounts a ON a.id = cc.id
		WHERE a.user_id = $1
		  AND e.type = 'DEBIT'
		  AND NOT t.opening_balance
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("user card payments: %w", err)
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (CardPayment, error) {
		var p CardPayment
		err := row.Scan(&p.CardID, &p.Date, &p.Amount)
		return p, err
	})
}

// UpsertBillTx records the bill a card statement prints (one per card per statement date; re-importing
// updates it), replacing the engine's estimate for that cycle. Cards without settings (no credit_cards
// row) aren't tracked, so it does nothing for them.
func UpsertBillTx(ctx context.Context, tx pgx.Tx, cardID uuid.UUID, start, statementDate time.Time, dueDate *time.Time, total decimal.Decimal) error {
	// The estimate may have guessed the statement date a few days off (a weekend rule, a holiday).
	if _, err := tx.Exec(ctx, `
		DELETE FROM credit_card_statements
		WHERE credit_card_id = $1 AND status = $2 AND statement_date <> $3
		  AND statement_date BETWEEN $3::date - 10 AND $3::date + 10
	`, cardID, StatementEstimated, statementDate); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO credit_card_statements (credit_card_id, start_date, end_date, statement_date, due_date, total_amount, status)
		SELECT $1, $2, $3, $3, $4, $5, $6
		WHERE EXISTS (SELECT 1 FROM credit_cards WHERE id = $1)
		ON CONFLICT (credit_card_id, statement_date) DO UPDATE
		SET start_date = EXCLUDED.start_date, due_date = EXCLUDED.due_date,
		    total_amount = EXCLUDED.total_amount, status = EXCLUDED.status, updated_at = now()
	`, cardID, start, statementDate, dueDate, total, StatementOpen)
	return err
}
