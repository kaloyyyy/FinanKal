package loan

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/schedule"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidLoan     = errors.New("invalid loan")
	ErrAlreadyImported = errors.New("this loan was already imported")
)

const (
	feesAccount     = "Loan Fees"
	interestAccount = "Loan Interest"
	equityAccount   = "Opening Balance Equity"
)

// Loan dates are Philippine calendar days.
var manila = time.FixedZone("PHT", 8*60*60)

type Service struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewService(db *pgxpool.Pool, redis *redis.Client) *Service {
	return &Service{db: db, redis: redis}
}

// Import is a loan read from the lender's disclosure statement.
type Import struct {
	UserID        uuid.UUID
	AccountNumber string
	AccountName   string
	BookedOn      time.Time
	Amount        decimal.Decimal // lent
	Deducted      decimal.Decimal // fees and taxes taken from the proceeds
	DailyRate     decimal.Decimal
	SourceID      uuid.UUID // where payments come from; uuid.Nil when unknown
	ScheduleName  string
	Installments  []schedule.Installment
	FileSHA256    string
}

type Imported struct {
	AccountID   uuid.UUID
	AccountName string
}

// Import creates the loan account, books the money lent and the deductions, and records its installments.
// shortcut: the proceeds are booked against Opening Balance Equity (the first loan's landed in an untracked
// account); take a proceeds account in the request when a loan pays into a tracked one.
func (s *Service) Import(ctx context.Context, in Import) (Imported, error) {
	if err := validate(in); err != nil {
		return Imported{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Imported{}, err
	}
	defer tx.Rollback(ctx)

	var seen bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM loans l JOIN accounts a ON a.id = l.account_id
		               WHERE l.file_sha256 = $2 OR (a.user_id = $1 AND a.account_number = $3))
	`, in.UserID, in.FileSHA256, in.AccountNumber).Scan(&seen); err != nil {
		return Imported{}, fmt.Errorf("check loan: %w", err)
	}
	if seen {
		return Imported{}, ErrAlreadyImported
	}
	if in.SourceID != uuid.Nil {
		var ok bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM accounts WHERE id = $1 AND user_id = $2 AND type = 'ASSET')`,
			in.SourceID, in.UserID).Scan(&ok); err != nil {
			return Imported{}, err
		}
		if !ok {
			return Imported{}, fmt.Errorf("%w: payments must come from one of your bank or wallet accounts", ErrInvalidLoan)
		}
	}

	// An account already holding the number (made by hand, say) becomes the loan if it's a liability.
	var loanID uuid.UUID
	var loanType string
	err = tx.QueryRow(ctx, `SELECT id, type FROM accounts WHERE user_id = $1 AND account_number = $2 FOR UPDATE`,
		in.UserID, in.AccountNumber).Scan(&loanID, &loanType)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		if err := tx.QueryRow(ctx, `INSERT INTO accounts (user_id, name, type, account_number) VALUES ($1, $2, 'LIABILITY', $3) RETURNING id`,
			in.UserID, in.AccountName, in.AccountNumber).Scan(&loanID); err != nil {
			return Imported{}, fmt.Errorf("create loan account: %w", err)
		}
	case err != nil:
		return Imported{}, err
	case loanType != "LIABILITY":
		return Imported{}, fmt.Errorf("%w: account ••%s is already one of your %s accounts", ErrInvalidLoan, last4(in.AccountNumber), loanType)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO loans (account_id, booked_on, daily_rate, principal_on, principal, file_sha256) VALUES ($1, $2, $3, $2, $4, $5)
	`, loanID, in.BookedOn, in.DailyRate, in.Amount, in.FileSHA256); err != nil {
		return Imported{}, fmt.Errorf("record loan: %w", err)
	}

	equityID, err := ensureAccount(ctx, tx, in.UserID, "EQUITY", equityAccount)
	if err != nil {
		return Imported{}, err
	}
	at := dayStart(in.BookedOn)
	if err := book(ctx, tx, in.UserID, in.AccountName+" disbursement", at, equityID, loanID, in.Amount.Sub(in.Deducted)); err != nil {
		return Imported{}, err
	}
	if in.Deducted.IsPositive() {
		feesID, err := ensureAccount(ctx, tx, in.UserID, "EXPENSE", feesAccount)
		if err != nil {
			return Imported{}, err
		}
		if err := book(ctx, tx, in.UserID, in.AccountName+" fees and taxes deducted", at, feesID, loanID, in.Deducted); err != nil {
			return Imported{}, err
		}
	}

	// The same schedule CreateInstallments records; its pending occurrences are the loan's deadlines.
	var source *uuid.UUID
	if in.SourceID != uuid.Nil {
		source = &in.SourceID
	}
	first, last := in.Installments[0], in.Installments[len(in.Installments)-1]
	var scheduleID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO scheduled_transactions (user_id, name, schedule_type, transaction_type, amount, source_account_id,
			destination_account_id, start_date, end_date, frequency, day_of_month, amount_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id
	`, in.UserID, in.ScheduleName, schedule.ScheduleTypeRecurring, schedule.TransactionTypeTransfer, first.Amount, source, loanID,
		first.DueDate, last.DueDate, schedule.FrequencyMonthly, first.DueDate.Day(), schedule.AmountTypeFixed).Scan(&scheduleID); err != nil {
		return Imported{}, fmt.Errorf("create loan schedule: %w", err)
	}
	for i, inst := range in.Installments {
		if _, err := tx.Exec(ctx, `
			INSERT INTO scheduled_transaction_occurrences (scheduled_transaction_id, scheduled_date, installment_number, amount, status)
			VALUES ($1, $2, $3, $4, $5)
		`, scheduleID, inst.DueDate, i+1, inst.Amount, schedule.OccurrencePending); err != nil {
			return Imported{}, fmt.Errorf("create installment %d: %w", i+1, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Imported{}, err
	}
	cache.InvalidateTransaction(ctx, s.redis, in.UserID, loanID, equityID)
	return Imported{AccountID: loanID, AccountName: in.AccountName}, nil
}

func validate(in Import) error {
	switch {
	case in.UserID == uuid.Nil:
		return fmt.Errorf("%w: user is required", ErrInvalidLoan)
	case in.AccountNumber == "" || in.AccountName == "":
		return fmt.Errorf("%w: the loan's account number and name are required", ErrInvalidLoan)
	case !in.Amount.IsPositive():
		return fmt.Errorf("%w: the loan amount must be more than zero", ErrInvalidLoan)
	case in.Deducted.IsNegative() || !in.Deducted.LessThan(in.Amount):
		return fmt.Errorf("%w: deductions must be less than the loan amount", ErrInvalidLoan)
	case in.DailyRate.IsNegative():
		return fmt.Errorf("%w: the interest rate can't be negative", ErrInvalidLoan)
	case len(in.FileSHA256) != 64:
		return fmt.Errorf("%w: file hash must be a SHA-256 hex string", ErrInvalidLoan)
	case len(in.Installments) == 0 || len(in.Installments) > 360:
		return fmt.Errorf("%w: a loan needs 1 to 360 installments", ErrInvalidLoan)
	case in.ScheduleName == "":
		return fmt.Errorf("%w: schedule name is required", ErrInvalidLoan)
	}
	for i, inst := range in.Installments {
		if !inst.Amount.IsPositive() || !inst.DueDate.After(in.BookedOn) || (i > 0 && !inst.DueDate.After(in.Installments[i-1].DueDate)) {
			return fmt.Errorf("%w: installment %d needs a positive amount and a due date after the one before", ErrInvalidLoan, i+1)
		}
	}
	return nil
}

type loanRow struct {
	id          uuid.UUID
	name        string
	bookedOn    time.Time
	rate        decimal.Decimal
	principalOn time.Time
	principal   decimal.Decimal
}

type Details struct {
	AccountID   uuid.UUID
	Name        string
	BookedOn    time.Time
	DailyRate   decimal.Decimal
	PrincipalOn time.Time
	Principal   decimal.Decimal
}

func (s *Service) Get(ctx context.Context, userID, accountID uuid.UUID) (Details, error) {
	var d Details
	err := s.db.QueryRow(ctx, `SELECT a.id,a.name,l.booked_on,l.daily_rate,l.principal_on,l.principal FROM loans l JOIN accounts a ON a.id=l.account_id WHERE a.user_id=$1 AND a.id=$2`, userID, accountID).Scan(&d.AccountID, &d.Name, &d.BookedOn, &d.DailyRate, &d.PrincipalOn, &d.Principal)
	return d, err
}

func (s *Service) Update(ctx context.Context, userID uuid.UUID, d Details) error {
	if d.AccountID == uuid.Nil || strings.TrimSpace(d.Name) == "" || !d.Principal.IsPositive() || d.DailyRate.IsNegative() || d.BookedOn.IsZero() || d.PrincipalOn.IsZero() || d.PrincipalOn.Before(d.BookedOn) {
		return fmt.Errorf("%w: provide a name, positive principal, nonnegative rate, and valid dates", ErrInvalidLoan)
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	res, err := tx.Exec(ctx, `UPDATE loans l SET booked_on=$3,daily_rate=$4,principal_on=$5,principal=$6 FROM accounts a WHERE l.account_id=$1 AND a.id=l.account_id AND a.user_id=$2`, d.AccountID, userID, d.BookedOn, d.DailyRate, d.PrincipalOn, d.Principal)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET name=$3 WHERE id=$1 AND user_id=$2`, d.AccountID, userID, strings.TrimSpace(d.Name)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	cache.InvalidateTransaction(ctx, s.redis, userID, d.AccountID)
	return nil
}

// Delete preserves the account and ledger; it removes only loan metadata and cancels future scheduled payments.
func (s *Service) Delete(ctx context.Context, userID, accountID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM loans l JOIN accounts a ON a.id=l.account_id WHERE l.account_id=$1 AND a.user_id=$2)`, accountID, userID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	if _, err := tx.Exec(ctx, `UPDATE scheduled_transaction_occurrences o SET status='SKIPPED' FROM scheduled_transactions s WHERE o.scheduled_transaction_id=s.id AND s.destination_account_id=$1 AND s.user_id=$2 AND o.status='PENDING'`, accountID, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE scheduled_transactions SET active=false,updated_at=now() WHERE destination_account_id=$1 AND user_id=$2`, accountID, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM loans WHERE account_id=$1`, accountID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	cache.InvalidateTransaction(ctx, s.redis, userID, accountID)
	return nil
}

// Accrue brings each of the user's loans up to today: the catch-up from the ledger to the principal the
// lender reported (when one was set), then the interest of every due date since. Nothing is written when
// the result is what's already booked.
// shortcut: runs on every payables read (like card estimates); move to the import path if reads get hot.
func (s *Service) Accrue(ctx context.Context, userID uuid.UUID, today time.Time) error {
	rows, err := s.db.Query(ctx, `
		SELECT l.account_id, a.name, l.booked_on, l.daily_rate, l.principal_on, l.principal
		FROM loans l JOIN accounts a ON a.id = l.account_id WHERE a.user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("list loans: %w", err)
	}
	loans, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (loanRow, error) {
		var l loanRow
		err := r.Scan(&l.id, &l.name, &l.bookedOn, &l.rate, &l.principalOn, &l.principal)
		return l, err
	})
	if err != nil {
		return err
	}
	for _, l := range loans {
		if err := s.accrue(ctx, userID, l, today); err != nil {
			return fmt.Errorf("accrue %s: %w", l.name, err)
		}
	}
	return nil
}

// booking is a loan charge as the ledger holds it: positive raises what's owed.
type booking struct {
	day         time.Time
	amount      decimal.Decimal
	description string
}

func (s *Service) accrue(ctx context.Context, userID uuid.UUID, l loanRow, today time.Time) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT 1 FROM loans WHERE account_id = $1 FOR UPDATE`, l.id); err != nil {
		return err
	}

	var want []booking
	if l.principalOn.After(l.bookedOn) {
		// What the ledger says was owed at the end of that day, without the engine's own charges.
		var owed decimal.Decimal
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(SUM(CASE WHEN e.type = 'CREDIT' THEN e.amount ELSE -e.amount END), 0)
			FROM entries e JOIN transactions t ON t.id = e.transaction_id
			WHERE e.account_id = $1 AND t.created_at < $2
			  AND t.id NOT IN (SELECT transaction_id FROM loan_charges WHERE account_id = $1)
		`, l.id, dayStart(l.principalOn.AddDate(0, 0, 1))).Scan(&owed); err != nil {
			return err
		}
		if gap := l.principal.Sub(owed); !gap.IsZero() {
			want = append(want, booking{l.principalOn, gap,
				fmt.Sprintf("%s interest and penalties to %s, per lender", l.name, l.principalOn.Format("2 Jan 2006"))})
		}
	}

	dueRows, err := tx.Query(ctx, `
		SELECT DISTINCT o.scheduled_date FROM scheduled_transaction_occurrences o
		JOIN scheduled_transactions s ON s.id = o.scheduled_transaction_id
		WHERE s.destination_account_id = $1 ORDER BY 1
	`, l.id)
	if err != nil {
		return err
	}
	dues, err := pgx.CollectRows(dueRows, pgx.RowTo[time.Time])
	if err != nil {
		return err
	}
	payments, err := loanPayments(ctx, tx, l.id)
	if err != nil {
		return err
	}
	for _, c := range Charges(l.principal, l.principalOn, dues, payments, l.rate, today) {
		want = append(want, booking{c.Due, c.Amount, fmt.Sprintf("%s interest to %s", l.name, c.Due.Format("2 Jan 2006"))})
	}

	have, err := booked(ctx, tx, l.id)
	if err != nil {
		return err
	}
	if slices.EqualFunc(have, want, func(a, b booking) bool {
		return a.day.Equal(b.day) && a.amount.Equal(b.amount) && a.description == b.description
	}) {
		return nil
	}

	if _, err := tx.Exec(ctx, `DELETE FROM transactions WHERE id IN (SELECT transaction_id FROM loan_charges WHERE account_id = $1)`, l.id); err != nil {
		return fmt.Errorf("clear loan charges: %w", err)
	}
	interestID, err := ensureAccount(ctx, tx, userID, "EXPENSE", interestAccount)
	if err != nil {
		return err
	}
	for _, b := range want {
		debit, credit := interestID, l.id
		if b.amount.IsNegative() {
			debit, credit = l.id, interestID
		}
		var txID uuid.UUID
		if err := tx.QueryRow(ctx, `INSERT INTO transactions (user_id, description, created_at) VALUES ($1, $2, $3) RETURNING id`,
			userID, b.description, dayStart(b.day)).Scan(&txID); err != nil {
			return err
		}
		if err := entries(ctx, tx, txID, debit, credit, b.amount.Abs()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO loan_charges (transaction_id, account_id) VALUES ($1, $2)`, txID, l.id); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	cache.InvalidateTransaction(ctx, s.redis, userID, l.id, interestID)
	return nil
}

// loanPayments is the money that reached the loan (its DEBIT entries), oldest first, by Manila calendar day.
func loanPayments(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, loanID uuid.UUID) ([]Payment, error) {
	rows, err := q.Query(ctx, `
		SELECT (t.created_at AT TIME ZONE 'Asia/Manila')::date, e.amount
		FROM entries e JOIN transactions t ON t.id = e.transaction_id
		WHERE e.account_id = $1 AND e.type = 'DEBIT'
		  AND t.id NOT IN (SELECT transaction_id FROM loan_charges WHERE account_id = $1)
		ORDER BY 1
	`, loanID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Payment, error) {
		var p Payment
		err := r.Scan(&p.Date, &p.Amount)
		return p, err
	})
}

// booked is the loan's current charges, in the order Accrue books them (catch-up first, then by due date).
func booked(ctx context.Context, tx pgx.Tx, loanID uuid.UUID) ([]booking, error) {
	rows, err := tx.Query(ctx, `
		SELECT (t.created_at AT TIME ZONE 'Asia/Manila')::date, CASE WHEN e.type = 'CREDIT' THEN e.amount ELSE -e.amount END, t.description
		FROM loan_charges c JOIN transactions t ON t.id = c.transaction_id
		JOIN entries e ON e.transaction_id = t.id AND e.account_id = c.account_id
		WHERE c.account_id = $1
		ORDER BY t.description NOT LIKE '%per lender', t.created_at
	`, loanID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (booking, error) {
		var b booking
		err := r.Scan(&b.day, &b.amount, &b.description)
		return b, err
	})
}

// Settle takes paid installments off a loan's payables: payments to the loan cover its installments
// oldest first, and a partly covered one shows what's left of it. A loan that's paid off has none left.
// When the lender's app says installments are paid through a date (paid_through), those are taken as
// paid, and only payments after that date cover later installments.
func (s *Service) Settle(ctx context.Context, userID uuid.UUID, payables []schedule.Payable) ([]schedule.Payable, error) {
	rows, err := s.db.Query(ctx, `
		SELECT l.account_id, COALESCE(l.paid_through, l.booked_on),
		       COALESCE(SUM(e.amount) FILTER (WHERE e.type = 'DEBIT' AND c.transaction_id IS NULL
		           AND (l.paid_through IS NULL OR (t.created_at AT TIME ZONE 'Asia/Manila')::date > l.paid_through)), 0),
		       COALESCE(SUM(CASE WHEN e.type = 'CREDIT' THEN e.amount ELSE -e.amount END), 0)
		FROM loans l JOIN accounts a ON a.id = l.account_id
		LEFT JOIN entries e ON e.account_id = l.account_id
		LEFT JOIN transactions t ON t.id = e.transaction_id
		LEFT JOIN loan_charges c ON c.transaction_id = e.transaction_id
		WHERE a.user_id = $1
		GROUP BY l.account_id, l.paid_through, l.booked_on
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("loan payments: %w", err)
	}
	paid := map[uuid.UUID]decimal.Decimal{}
	owed := map[uuid.UUID]decimal.Decimal{}
	settledBy := map[uuid.UUID]time.Time{}
	for rows.Next() {
		var id uuid.UUID
		var pinned time.Time
		var p, o decimal.Decimal
		if err := rows.Scan(&id, &pinned, &p, &o); err != nil {
			rows.Close()
			return nil, err
		}
		paid[id], owed[id], settledBy[id] = p, o, pinned
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]schedule.Payable, 0, len(payables))
	for _, p := range payables {
		left, isLoan := paid[p.AccountID]
		if !isLoan {
			out = append(out, p)
			continue
		}
		if !owed[p.AccountID].IsPositive() || !p.DueDate.After(settledBy[p.AccountID]) {
			continue
		}
		covered := decimal.Min(left, p.Amount)
		paid[p.AccountID] = left.Sub(covered)
		if covered.Equal(p.Amount) {
			continue
		}
		p.Amount = p.Amount.Sub(covered)
		out = append(out, p)
	}
	return out, nil
}

func ensureAccount(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountType, name string) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM accounts WHERE user_id = $1 AND type = $2 AND name = $3 ORDER BY created_at, id LIMIT 1`,
		userID, accountType, name).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, `INSERT INTO accounts (user_id, name, type) VALUES ($1, $2, $3) RETURNING id`, userID, name, accountType).Scan(&id)
	}
	return id, err
}

// book records one transaction moving amount: DEBIT debit / CREDIT credit.
func book(ctx context.Context, tx pgx.Tx, userID uuid.UUID, description string, at time.Time, debit, credit uuid.UUID, amount decimal.Decimal) error {
	var txID uuid.UUID
	if err := tx.QueryRow(ctx, `INSERT INTO transactions (user_id, description, created_at) VALUES ($1, $2, $3) RETURNING id`,
		userID, description, at).Scan(&txID); err != nil {
		return err
	}
	return entries(ctx, tx, txID, debit, credit, amount)
}

func entries(ctx context.Context, tx pgx.Tx, txID, debit, credit uuid.UUID, amount decimal.Decimal) error {
	_, err := tx.Exec(ctx, `INSERT INTO entries (transaction_id, account_id, amount, type) VALUES ($1, $2, $4, 'DEBIT'), ($1, $3, $4, 'CREDIT')`,
		txID, debit, credit, amount)
	return err
}

func dayStart(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, manila)
}

func last4(number string) string {
	if len(number) <= 4 {
		return number
	}
	return number[len(number)-4:]
}
