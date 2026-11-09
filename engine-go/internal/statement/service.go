// Package statement imports bank statements into the ledger (see docs/adr/0001-bank-statement-import.md).
package statement

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/credit"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidStatement = errors.New("invalid statement")
	ErrAlreadyImported  = errors.New("this statement, or another one for the same period, was already imported into this account")
)

const (
	defaultExpenseAccount = "Outgoing Transfers"
	defaultIncomeAccount  = "Incoming Transfers"
	defaultEquityAccount  = "Opening Balance Equity"
	openingBalanceName    = "Opening balance"
	// Withdrawals (BankStatementParser's "Cash" category) go into the user's Cash account and are
	// spent from it the same day, into this spending account.
	cashAccountName  = "Cash"
	cashSpendingName = "Cash Spending"

	// How many days apart two banks may date the same transfer (sent late at night, received next day).
	transferDays = 2
	// Where the fee goes when the sender paid more than arrived (GCash and InstaPay charge ₱10–15).
	transferFeeAccount = "Transfer Fees"
)

var (
	minTransferFee = decimal.NewFromInt(10)
	maxTransferFee = decimal.NewFromInt(15)
)

// Statement dates are Philippine calendar days; Spring's TransactionMapper uses the same zone.
var manila = time.FixedZone("PHT", 8*60*60)

// Row is one line of a statement. Date is a calendar date (time of day and zone are ignored).
type Row struct {
	Date        time.Time
	Description string
	Amount      decimal.Decimal
	MoneyOut    bool
	Category    string // income / expense account name; "" = Incoming / Outgoing Transfers
}

// Statement is a parsed statement for one bank / wallet account. Period dates are calendar dates.
// Opening and closing balances are optional because not every bank prints them.
//
// AccountID is optional: without it the account is found by AccountNumber, or created as AccountName.
// With it, the account must belong to UserID, and a statement number is saved on it (or must match it).
type Statement struct {
	UserID         uuid.UUID
	AccountID      uuid.UUID
	AccountNumber  string
	AccountName    string
	AccountType    string // ASSET (bank / wallet, the default) or CREDIT_CARD
	Bill           *Bill  // a credit card statement's bill: what's due and when; nil for a bank statement or a card history
	Bank           string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	OpeningBalance decimal.NullDecimal
	ClosingBalance decimal.NullDecimal
	FileSHA256     string
	Rows           []Row
}

// Bill is what a credit card statement says is due. DueDate is nil when it isn't printed; the card's rule decides then.
type Bill struct {
	StatementDate time.Time
	DueDate       *time.Time
	TotalDue      decimal.Decimal
}

type Result struct {
	StatementID uuid.UUID
	AccountID   uuid.UUID
	AccountName string
	Imported    int
	Duplicates  []string // rows left out because the account already had them
	Transfers   []string // rows joined with the other end of the transfer on another of the user's accounts
	Warnings    []string
}

type Service struct {
	ledgerRepo *ledger.LedgerRepository
	redis      *redis.Client
}

func NewService(ledgerRepo *ledger.LedgerRepository, redis *redis.Client) *Service {
	return &Service{ledgerRepo: ledgerRepo, redis: redis}
}

// Import records the statement and its rows, recalculates the account's single opening balance and
// reconciles every statement of the account, all in one database transaction.
func (s *Service) Import(ctx context.Context, st Statement) (Result, error) {
	if err := validate(st); err != nil {
		return Result{}, err
	}

	tx, err := s.ledgerRepo.BeginTx(ctx)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback(ctx)

	acct, err := s.resolveAccount(ctx, tx, st)
	if err != nil {
		return Result{}, err
	}
	// Bank / wallet and credit card statements both work as debits − credits: on a card, a charge is money
	// out (owed goes up) and a payment is money in. A statement must go into an account of its own kind.
	if acct.Type != string(ledger.ASSET) && acct.Type != string(ledger.CREDIT_CARD) {
		return Result{}, fmt.Errorf("%w: statements can only be imported into a bank, wallet or credit card account", ErrInvalidStatement)
	}
	if acct.Type != st.accountType() {
		return Result{}, fmt.Errorf("%w: this is a %s statement, but %s is a %s account",
			ErrInvalidStatement, kindName(st.accountType()), acct.Name, kindName(acct.Type))
	}
	st.AccountID = acct.ID
	userID := acct.UserID

	statementID, err := insertStatement(ctx, tx, st)
	if err != nil {
		return Result{}, err
	}

	// Read before inserting, so identical rows within this statement (two ₱100 transfers on one day) all import.
	existing, err := existingRows(ctx, tx, st.AccountID, st.PeriodStart, st.PeriodEnd)
	if err != nil {
		return Result{}, err
	}

	touched := []uuid.UUID{st.AccountID}
	counters := map[string]uuid.UUID{}   // income / expense accounts by type and name, looked up once each
	cashAccounts := map[uuid.UUID]bool{} // the user's own Cash account, once a withdrawal has found it
	var duplicates, transfers []string

	for _, row := range st.Rows {
		// An overlapping statement, a matched transfer or a hand-entered transaction already has this movement.
		if takeExisting(existing, row) {
			direction := "in"
			if row.MoneyOut {
				direction = "out"
			}
			duplicates = append(duplicates, fmt.Sprintf("%s · %s · ₱%s %s",
				row.Date.Format("2 Jan 2006"), row.Description, row.Amount.StringFixed(2), direction))
			continue
		}

		// Money moved between two of the user's own accounts: one transfer, not spending here and income there.
		// shortcut: the transfer keeps the first-imported bank's date and statement; when the banks date it on
		// different days across a month end, that month's balance check on the other account warns.
		// A categorised row (salary, say) is never one end of a transfer. A card purchase can be: rent paid
		// through the card that comes back to a bank account is money moved from the card to that bank.
		var match *transferCandidate
		if row.Category == "" {
			if match, err = counterpart(ctx, tx, userID, st.AccountID, st.Bank, row); err != nil {
				return Result{}, err
			}
		}
		if match != nil {
			if err := moveEntry(ctx, tx, match.counterEntryID, st.AccountID, row.Amount); err != nil {
				return Result{}, err
			}
			touched = append(touched, match.otherAccountID, match.counterAccountID)
			direction := "from"
			if row.MoneyOut {
				direction = "to"
			}
			note := fmt.Sprintf("%s · ₱%s %s %s",
				row.Date.Format("2 Jan 2006"), row.Amount.StringFixed(2), direction, match.otherName)
			// The sender paid more than arrived: the difference is a fee, in the same transaction so it balances.
			if match.fee.IsPositive() {
				feeID, err := s.ensureAccount(ctx, tx, userID, ledger.EXPENSE, transferFeeAccount)
				if err != nil {
					return Result{}, err
				}
				if _, err := s.ledgerRepo.InsertEntryTx(ctx, tx, match.transactionID, feeID, match.fee, ledger.DEBIT); err != nil {
					return Result{}, err
				}
				touched = append(touched, feeID)
				note += fmt.Sprintf(" (₱%s fee)", match.fee.StringFixed(2))
			}
			transfers = append(transfers, note)
			continue
		}

		accountType, name := ledger.INCOME, defaultIncomeAccount
		if row.MoneyOut {
			accountType, name = ledger.EXPENSE, defaultExpenseAccount
		}
		if row.Category != "" {
			name = row.Category
		}
		key := string(accountType) + "|" + name
		counter, ok := counters[key]
		if !ok {
			// A category naming one of the user's own money accounts ("Cash") or loans ("Maya Personal Loan")
			// moves the money there: an ATM withdrawal goes into Cash, a loan payment pays the loan down.
			own := uuid.Nil
			if row.Category != "" {
				if own, err = accountNamed(ctx, tx, userID, string(ledger.ASSET), name); err != nil {
					return Result{}, err
				}
				if own == uuid.Nil {
					if own, err = accountNamed(ctx, tx, userID, string(ledger.LIABILITY), name); err != nil {
						return Result{}, err
					}
				}
			}
			if counter = own; counter == uuid.Nil {
				if counter, err = s.ensureAccount(ctx, tx, userID, accountType, name); err != nil {
					return Result{}, err
				}
			} else if name == cashAccountName {
				cashAccounts[counter] = true
			}
			counters[key] = counter
			touched = append(touched, counter)
		}

		txID, err := insertTransaction(ctx, tx, userID, row.Description, dayStart(row.Date), &statementID, false)
		if err != nil {
			return Result{}, err
		}
		// Money out: DEBIT expense / CREDIT bank. Money in: DEBIT bank / CREDIT income.
		debit, credit := st.AccountID, counter
		if row.MoneyOut {
			debit, credit = counter, st.AccountID
		}
		if err := s.insertPair(ctx, tx, txID, debit, credit, row.Amount); err != nil {
			return Result{}, err
		}

		// Withdrawn cash counts as spent the same day: Cash → Cash Spending, so the Cash account stays at
		// zero and net position doesn't count money that's already in the user's pocket and gone.
		if row.MoneyOut && row.Category == cashAccountName && cashAccounts[counter] {
			spending, ok := counters["cash-spending"]
			if !ok {
				if spending, err = s.ensureAccount(ctx, tx, userID, ledger.EXPENSE, cashSpendingName); err != nil {
					return Result{}, err
				}
				counters["cash-spending"] = spending
				touched = append(touched, spending)
			}
			spentID, err := insertTransaction(ctx, tx, userID, "Cash spent: "+row.Description, dayStart(row.Date), &statementID, false)
			if err != nil {
				return Result{}, err
			}
			if err := s.insertPair(ctx, tx, spentID, spending, counter, row.Amount); err != nil {
				return Result{}, err
			}
		}
	}

	// A card statement's bill: its due date goes into payables, and payments reaching the card settle it.
	if st.Bill != nil && acct.Type == string(ledger.CREDIT_CARD) {
		if err := credit.UpsertBillTx(ctx, tx, st.AccountID, st.PeriodStart, st.Bill.StatementDate, st.Bill.DueDate, st.Bill.TotalDue); err != nil {
			return Result{}, err
		}
	}

	equityID, err := s.recalculateOpeningBalance(ctx, tx, userID, st.AccountID)
	if err != nil {
		return Result{}, err
	}
	if equityID != uuid.Nil {
		touched = append(touched, equityID)
	}

	warnings, err := reconcile(ctx, tx, st.AccountID)
	if err != nil {
		return Result{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Result{}, err
	}
	cache.InvalidateTransaction(ctx, s.redis, userID, touched...)

	return Result{StatementID: statementID, AccountID: acct.ID, AccountName: acct.Name, Imported: len(st.Rows) - len(duplicates) - len(transfers),
		Duplicates: duplicates, Transfers: transfers, Warnings: warnings}, nil
}

// List returns the user's imported statements, oldest period first.
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Summary, error) {
	tx, err := s.ledgerRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	return userStatements(ctx, tx, userID)
}

// resolveAccount finds the statement's account (by id, else by number) and locks it, creating it when
// no account has the number yet. A number on the statement is remembered on the account it went into.
func (s *Service) resolveAccount(ctx context.Context, tx pgx.Tx, st Statement) (account, error) {
	if st.AccountID == uuid.Nil {
		if st.AccountNumber == "" {
			return account{}, fmt.Errorf("%w: no account number was found on the statement; choose the account it belongs to", ErrInvalidStatement)
		}
		acct, found, err := lockAccountByNumber(ctx, tx, st.UserID, st.AccountNumber)
		if err != nil || found {
			return acct, err
		}
		id, err := s.ledgerRepo.CreateAccountTx(ctx, tx, st.UserID, st.AccountName, ledger.AccountType(st.accountType()))
		if err != nil {
			return account{}, err
		}
		if _, err := setAccountNumber(ctx, tx, id, st.AccountNumber); err != nil {
			return account{}, err
		}
		return account{ID: id, UserID: st.UserID, Name: st.AccountName, Type: st.accountType(), Number: st.AccountNumber}, nil
	}

	acct, err := lockAccount(ctx, tx, st.AccountID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && acct.UserID != st.UserID) {
		return account{}, fmt.Errorf("%w: the chosen account wasn't found", ErrInvalidStatement)
	}
	if err != nil || st.AccountNumber == "" {
		return acct, err
	}
	switch {
	case acct.Number == "":
		ok, err := setAccountNumber(ctx, tx, acct.ID, st.AccountNumber)
		if err != nil {
			return account{}, err
		}
		if !ok {
			return account{}, fmt.Errorf("%w: another of your accounts already has account number ••%s; choose that one or let FinanKal pick",
				ErrInvalidStatement, last4(st.AccountNumber))
		}
		acct.Number = st.AccountNumber
	case acct.Number != st.AccountNumber:
		return account{}, fmt.Errorf("%w: this statement is for account ••%s, but %s is ••%s",
			ErrInvalidStatement, last4(st.AccountNumber), acct.Name, last4(acct.Number))
	}
	return acct, nil
}

func last4(number string) string {
	if len(number) <= 4 {
		return number
	}
	return number[len(number)-4:]
}

// recalculateOpeningBalance keeps exactly one opening balance for the account, dated at its earliest
// statement, so the ledger starts where that statement starts no matter which month was imported first.
// Returns the equity account it used, or uuid.Nil when no opening balance is needed.
func (s *Service) recalculateOpeningBalance(ctx context.Context, tx pgx.Tx, userID, accountID uuid.UUID) (uuid.UUID, error) {
	if err := deleteOpeningBalances(ctx, tx, accountID); err != nil {
		return uuid.Nil, err
	}

	all, err := statements(ctx, tx, accountID)
	if err != nil || len(all) == 0 || !all[0].OpeningBalance.Valid {
		return uuid.Nil, err
	}
	earliest := all[0]

	before, err := balanceBefore(ctx, tx, accountID, dayStart(earliest.PeriodStart))
	if err != nil {
		return uuid.Nil, err
	}
	gap := earliest.OpeningBalance.Decimal.Sub(before)
	if gap.IsZero() {
		return uuid.Nil, nil
	}

	equityID, err := s.ensureAccount(ctx, tx, userID, ledger.EQUITY, defaultEquityAccount)
	if err != nil {
		return uuid.Nil, err
	}
	txID, err := insertTransaction(ctx, tx, userID, openingBalanceName, dayStart(earliest.PeriodStart), nil, true)
	if err != nil {
		return uuid.Nil, err
	}
	// Raising the bank balance is DEBIT bank / CREDIT equity; lowering it is the reverse.
	debit, credit := accountID, equityID
	if gap.IsNegative() {
		debit, credit = equityID, accountID
	}
	return equityID, s.insertPair(ctx, tx, txID, debit, credit, gap.Abs())
}

// reconcile checks every statement of the account against the ledger and describes each mismatch.
// Mismatches are reported, never adjusted away: they mean a missing statement, a manual entry, or a misread row.
func reconcile(ctx context.Context, tx pgx.Tx, accountID uuid.UUID) ([]string, error) {
	all, err := statements(ctx, tx, accountID)
	if err != nil {
		return nil, err
	}

	var warnings []string
	for _, st := range all {
		period := st.PeriodStart.Format("2 Jan") + "–" + st.PeriodEnd.Format("2 Jan 2006")

		if st.OpeningBalance.Valid {
			got, err := balanceBefore(ctx, tx, accountID, dayStart(st.PeriodStart))
			if err != nil {
				return nil, err
			}
			if !got.Equal(st.OpeningBalance.Decimal) {
				warnings = append(warnings, fmt.Sprintf(
					"%s: FinanKal has ₱%s at the start, but the statement starts at ₱%s. An earlier statement or transaction may be missing.",
					period, got.StringFixed(2), st.OpeningBalance.Decimal.StringFixed(2)))
			}
		}
		if st.ClosingBalance.Valid {
			got, err := balanceBefore(ctx, tx, accountID, dayStart(st.PeriodEnd.AddDate(0, 0, 1)))
			if err != nil {
				return nil, err
			}
			if !got.Equal(st.ClosingBalance.Decimal) {
				warnings = append(warnings, fmt.Sprintf(
					"%s: FinanKal ends at ₱%s, but the statement ends at ₱%s. Some rows may have been misread or entered twice.",
					period, got.StringFixed(2), st.ClosingBalance.Decimal.StringFixed(2)))
			}
		}
	}
	return warnings, nil
}

// ensureAccount returns the user's account of the type with that name, creating it if there is none.
// Imports keep to their own accounts, so a "Salary" income account isn't filled with every transfer.
func (s *Service) ensureAccount(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountType ledger.AccountType, name string) (uuid.UUID, error) {
	id, err := accountNamed(ctx, tx, userID, string(accountType), name)
	if err != nil || id != uuid.Nil {
		return id, err
	}
	return s.ledgerRepo.CreateAccountTx(ctx, tx, userID, name, accountType)
}

func (s *Service) insertPair(ctx context.Context, tx pgx.Tx, txID, debit, credit uuid.UUID, amount decimal.Decimal) error {
	if _, err := s.ledgerRepo.InsertEntryTx(ctx, tx, txID, debit, amount, ledger.DEBIT); err != nil {
		return err
	}
	_, err := s.ledgerRepo.InsertEntryTx(ctx, tx, txID, credit, amount, ledger.CREDIT)
	return err
}

func validate(st Statement) error {
	switch {
	case st.UserID == uuid.Nil:
		return fmt.Errorf("%w: user is required", ErrInvalidStatement)
	case st.AccountType != "" && st.AccountType != string(ledger.ASSET) && st.AccountType != string(ledger.CREDIT_CARD):
		return fmt.Errorf("%w: account type must be ASSET or CREDIT_CARD", ErrInvalidStatement)
	case st.AccountID == uuid.Nil && st.AccountNumber != "" && st.AccountName == "":
		return fmt.Errorf("%w: a name is needed in case the account has to be created", ErrInvalidStatement)
	case st.PeriodEnd.Before(st.PeriodStart):
		return fmt.Errorf("%w: the period ends before it starts", ErrInvalidStatement)
	case len(st.FileSHA256) != 64:
		return fmt.Errorf("%w: file hash must be a SHA-256 hex string", ErrInvalidStatement)
	case st.Bank == "":
		return fmt.Errorf("%w: bank is required", ErrInvalidStatement)
	}
	for i, row := range st.Rows {
		if !row.Amount.IsPositive() {
			return fmt.Errorf("%w: row %d has a non-positive amount", ErrInvalidStatement, i+1)
		}
		if row.Description == "" {
			return fmt.Errorf("%w: row %d has no description", ErrInvalidStatement, i+1)
		}
	}
	return nil
}

// takeExisting reports whether the account already has row, and uses that movement up so the next identical
// row (two ₱100 transfers on one day) isn't matched to it too. Same day, amount and direction; a transfer
// may be up to transferDays off, because it carries the date the other bank gave it.
func takeExisting(existing []*existingRow, row Row) bool {
	var best *existingRow
	bestGap := transferDays + 1
	for _, e := range existing {
		if e.used || e.moneyOut != row.MoneyOut || !e.amount.Equal(row.Amount) {
			continue
		}
		gap := daysApart(e.day, row.Date)
		if gap == 0 {
			best = e
			break
		}
		if e.transfer && gap < bestGap {
			best, bestGap = e, gap
		}
	}
	if best == nil {
		return false
	}
	best.used = true
	return true
}

func daysApart(a, b time.Time) int {
	days := int(calendarDate(a).Sub(calendarDate(b)).Hours() / 24)
	if days < 0 {
		return -days
	}
	return days
}

// calendarDate drops the time and zone of a date-only value.
func calendarDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// accountType is the kind of account the statement is for; bank / wallet unless it says credit card.
func (st Statement) accountType() string {
	if st.AccountType == "" {
		return string(ledger.ASSET)
	}
	return st.AccountType
}

func kindName(accountType string) string {
	if accountType == string(ledger.CREDIT_CARD) {
		return "credit card"
	}
	if accountType == string(ledger.ASSET) {
		return "bank or wallet"
	}
	return strings.ToLower(strings.ReplaceAll(accountType, "_", " "))
}

// dayStart is midnight in Manila on the given calendar date.
func dayStart(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, manila)
}
