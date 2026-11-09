package statement

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

// stored is a statement already in the database, used for reconciliation.
type stored struct {
	PeriodStart    time.Time
	PeriodEnd      time.Time
	OpeningBalance decimal.NullDecimal
	ClosingBalance decimal.NullDecimal
}

type account struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	Type   string
	Number string // "" when not known yet
}

// Imports lock the account so imports into the same account run one after another
// (the opening balance depends on all of them).
const selectAccount = `SELECT id, user_id, name, type, COALESCE(account_number, '') FROM accounts `

func scanAccount(row pgx.Row) (account, error) {
	var a account
	err := row.Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Number)
	return a, err
}

func lockAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID) (account, error) {
	return scanAccount(tx.QueryRow(ctx, selectAccount+`WHERE id = $1 FOR UPDATE`, accountID))
}

// lockAccountByNumber returns found=false when the user has no account with that number.
func lockAccountByNumber(ctx context.Context, tx pgx.Tx, userID uuid.UUID, number string) (account, bool, error) {
	a, err := scanAccount(tx.QueryRow(ctx, selectAccount+`WHERE user_id = $1 AND account_number = $2 FOR UPDATE`, userID, number))
	if errors.Is(err, pgx.ErrNoRows) {
		return account{}, false, nil
	}
	return a, err == nil, err
}

// setAccountNumber returns ok=false when another of the user's accounts already has the number.
func setAccountNumber(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, number string) (bool, error) {
	_, err := tx.Exec(ctx, `UPDATE accounts SET account_number = $2 WHERE id = $1`, accountID, number)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return false, nil
	}
	return err == nil, err
}

func insertStatement(ctx context.Context, tx pgx.Tx, s Statement) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO bank_statements (
			account_id, bank, period_start, period_end, opening_balance, closing_balance, file_sha256
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
	`, s.AccountID, s.Bank, s.PeriodStart, s.PeriodEnd, s.OpeningBalance, s.ClosingBalance, s.FileSHA256).Scan(&id)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation: same file or same period
		return uuid.Nil, ErrAlreadyImported
	}
	return id, err
}

// accountNamed returns the user's oldest account of the type with that name, or uuid.Nil if there is none.
func accountNamed(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountType, name string) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT id FROM accounts
		WHERE user_id = $1 AND type = $2 AND name = $3
		ORDER BY created_at, id
		LIMIT 1
	`, userID, accountType, name).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, nil
	}
	return id, err
}

func insertTransaction(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
	description string,
	at time.Time,
	statementID *uuid.UUID,
	openingBalance bool,
) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO transactions (user_id, description, created_at, statement_id, opening_balance)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id
	`, userID, description, at, statementID, openingBalance).Scan(&id)
	return id, err
}

func deleteOpeningBalances(ctx context.Context, tx pgx.Tx, accountID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM transactions t
		WHERE t.opening_balance
		  AND EXISTS (SELECT 1 FROM entries e WHERE e.transaction_id = t.id AND e.account_id = $1)
	`, accountID)
	return err
}

// balanceBefore is the account's balance (debits − credits) from transactions dated before at.
func balanceBefore(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, at time.Time) (decimal.Decimal, error) {
	var balance decimal.Decimal
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(CASE WHEN e.type = 'DEBIT' THEN e.amount ELSE -e.amount END), 0)
		FROM entries e
		JOIN transactions t ON t.id = e.transaction_id
		WHERE e.account_id = $1
		  AND t.created_at < $2
	`, accountID, at).Scan(&balance)
	return balance, err
}

// existingRow is a movement already on the account. Money out is a CREDIT on the bank account.
type existingRow struct {
	day      time.Time // Manila calendar date
	amount   decimal.Decimal
	moneyOut bool
	transfer bool // the other side is another bank account: it may carry the other bank's date
	used     bool
}

// existingRows returns the account's movements (opening balances excluded) dated from start to end,
// widened by transferDays so a transfer dated by the other bank is still found.
// Rows from earlier statements, transfers matched from another account, and hand-entered ones all count.
func existingRows(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, start, end time.Time) ([]*existingRow, error) {
	rows, err := tx.Query(ctx, `
		SELECT (t.created_at AT TIME ZONE 'Asia/Manila')::date, e.amount, e.type = 'CREDIT',
		       EXISTS (
		           SELECT 1 FROM entries o JOIN accounts oa ON oa.id = o.account_id
		           WHERE o.transaction_id = t.id AND o.id <> e.id AND oa.type = 'ASSET'
		       )
		FROM entries e
		JOIN transactions t ON t.id = e.transaction_id
		WHERE e.account_id = $1
		  AND NOT t.opening_balance
		  AND t.created_at >= $2 AND t.created_at < $3
	`, accountID, dayStart(start.AddDate(0, 0, -transferDays)), dayStart(end.AddDate(0, 0, 1+transferDays)))
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (*existingRow, error) {
		e := &existingRow{}
		err := row.Scan(&e.day, &e.amount, &e.moneyOut, &e.transfer)
		return e, err
	})
}

// transferCandidate is the other bank's half of a transfer, still booked as income or spending.
type transferCandidate struct {
	transactionID    uuid.UUID
	counterEntryID   uuid.UUID // its Incoming / Outgoing Transfers entry, which moves to this account
	counterAccountID uuid.UUID
	otherAccountID   uuid.UUID
	otherName        string
	gap              int             // days between the two banks' dates
	fee              decimal.Decimal // what the sender paid on top of what arrived (GCash, InstaPay: ₱10–15)
	named            bool            // one side's description names the other side's bank
}

// counterpart finds the row on another of the user's bank accounts that is the other end of row:
// imported from a statement, going the other way, within transferDays, and still on Incoming / Outgoing
// Transfers (not matched yet, and not categorised as salary or the like). The amount must be the same, or
// the sender's ₱10–15 more (a transfer fee) when one description names the other bank ("Received GCash
// from MariBank…", "Transfer to GCash"). bank is this statement's bank.
// Returns nil when there is none, or when the two best are equally good.
func counterpart(ctx context.Context, tx pgx.Tx, userID, accountID uuid.UUID, bank string, row Row) (*transferCandidate, error) {
	otherSide := "DEBIT" // money leaving this account arrives in the other one
	if !row.MoneyOut {
		otherSide = "CREDIT"
	}
	day := calendarDate(row.Date)
	rows, err := tx.Query(ctx, `
		SELECT t.id, ce.id, ce.account_id, oa.id, oa.name,
		       abs((t.created_at AT TIME ZONE 'Asia/Manila')::date - $4::date),
		       oe.amount, t.description, bs.bank
		FROM transactions t
		JOIN bank_statements bs ON bs.id = t.statement_id
		JOIN entries oe ON oe.transaction_id = t.id
		JOIN accounts oa ON oa.id = oe.account_id
		JOIN entries ce ON ce.transaction_id = t.id AND ce.id <> oe.id
		JOIN accounts ca ON ca.id = ce.account_id
		WHERE oa.user_id = $1 AND oa.id <> $2
		  -- another bank or wallet, or a card (a payment to it, or money drawn from it into a bank)
		  AND oa.type IN ('ASSET', 'CREDIT_CARD')
		  AND NOT t.opening_balance
		  AND oe.amount BETWEEN $3::numeric - $10 AND $3::numeric + $10 AND oe.type = $5
		  AND ca.type IN ('INCOME', 'EXPENSE') AND ca.name IN ($8, $9) -- not a categorised row such as salary
		  AND (SELECT count(*) FROM entries x WHERE x.transaction_id = t.id) = 2
		  AND t.created_at >= $6 AND t.created_at < $7
	`, userID, accountID, row.Amount, day, otherSide,
		dayStart(day.AddDate(0, 0, -transferDays)), dayStart(day.AddDate(0, 0, transferDays+1)),
		defaultIncomeAccount, defaultExpenseAccount, maxTransferFee)
	if err != nil {
		return nil, err
	}
	type scanned struct {
		transferCandidate
		amount      decimal.Decimal
		description string
		otherBank   string
	}
	found, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (scanned, error) {
		var c scanned
		err := r.Scan(&c.transactionID, &c.counterEntryID, &c.counterAccountID, &c.otherAccountID, &c.otherName,
			&c.gap, &c.amount, &c.description, &c.otherBank)
		return c, err
	})
	if err != nil {
		return nil, err
	}
	var best []transferCandidate
	for _, c := range found {
		c.named = names(row.Description, c.otherBank) || names(c.description, bank)
		fee, ok := transferFee(row, c.amount, c.named)
		if !ok {
			continue
		}
		c.fee = fee
		best = append(best, c.transferCandidate)
	}
	sort.Slice(best, func(i, j int) bool { return better(best[i], best[j]) })
	if len(best) == 0 || (len(best) > 1 && !better(best[0], best[1])) {
		return nil, nil
	}
	return &best[0], nil
}

// better ranks transfer candidates: named first, then an exact amount, then the closest date.
func better(a, b transferCandidate) bool {
	if a.named != b.named {
		return a.named
	}
	if a.fee.IsZero() != b.fee.IsZero() {
		return a.fee.IsZero()
	}
	return a.gap < b.gap
}

// transferFee is what the sender paid on top of what arrived: zero for the same amount, ₱10–15 when the
// two rows name each other's bank. ok=false when other (the other bank's amount) can't be the other end.
func transferFee(row Row, other decimal.Decimal, named bool) (decimal.Decimal, bool) {
	fee := row.Amount.Sub(other) // this side sent
	if !row.MoneyOut {
		fee = other.Sub(row.Amount) // the other side sent
	}
	if fee.IsZero() {
		return fee, true
	}
	return fee, named && fee.GreaterThanOrEqual(minTransferFee) && fee.LessThanOrEqual(maxTransferFee)
}

// names reports whether a statement description mentions bank ("Received GCash from MariBank…" names
// MARIBANK). Case and spaces are ignored, so "Union Bank" names UNIONBANK.
// shortcut: matches the bank's short name only; add aliases ("Bank of the Philippine Islands") if real rows use them.
func names(description, bank string) bool {
	squash := func(s string) string { return strings.ToLower(strings.ReplaceAll(s, " ", "")) }
	b := squash(bank)
	return len(b) >= 3 && strings.Contains(squash(description), b)
}

// moveEntry turns the other bank's income / spending row into a transfer by moving its
// Incoming / Outgoing Transfers side onto this account, at this account's amount (a fee is booked apart).
func moveEntry(ctx context.Context, tx pgx.Tx, entryID, accountID uuid.UUID, amount decimal.Decimal) error {
	_, err := tx.Exec(ctx, `UPDATE entries SET account_id = $2, amount = $3 WHERE id = $1`, entryID, accountID, amount)
	return err
}

// Summary is an imported statement as listed to the user.
type Summary struct {
	ID          uuid.UUID
	AccountID   uuid.UUID
	Bank        string
	PeriodStart time.Time
	PeriodEnd   time.Time
	ImportedAt  time.Time
}

func userStatements(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]Summary, error) {
	rows, err := tx.Query(ctx, `
		SELECT b.id, b.account_id, b.bank, b.period_start, b.period_end, b.imported_at
		FROM bank_statements b
		JOIN accounts a ON a.id = b.account_id
		WHERE a.user_id = $1
		ORDER BY b.period_start, b.period_end
	`, userID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Summary, error) {
		var s Summary
		err := row.Scan(&s.ID, &s.AccountID, &s.Bank, &s.PeriodStart, &s.PeriodEnd, &s.ImportedAt)
		return s, err
	})
}

// statements returns the account's statements, oldest period first.
func statements(ctx context.Context, tx pgx.Tx, accountID uuid.UUID) ([]stored, error) {
	rows, err := tx.Query(ctx, `
		SELECT period_start, period_end, opening_balance, closing_balance
		FROM bank_statements
		WHERE account_id = $1
		ORDER BY period_start, period_end
	`, accountID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (stored, error) {
		var s stored
		err := row.Scan(&s.PeriodStart, &s.PeriodEnd, &s.OpeningBalance, &s.ClosingBalance)
		return s, err
	})
}
