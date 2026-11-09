package statement_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/kaloy/finankal/engine-go/internal/statement"
	"github.com/shopspring/decimal"
)

// Like the credit tests: point TEST_DATABASE_URL at a throwaway database loaded from db/schema.sql.

type fixture struct {
	t        *testing.T
	ctx      context.Context
	pool     *pgxpool.Pool
	ledger   *ledger.Service
	service  *statement.Service
	userID   uuid.UUID
	bankID   uuid.UUID
	fileSeed int
}

func setup(t *testing.T) *fixture {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	repo := ledger.NewLedgerRepository(pool)
	f := &fixture{t: t, ctx: ctx, pool: pool, ledger: ledger.NewService(repo, nil), service: statement.NewService(repo, nil), userID: uuid.New()}

	if _, err := pool.Exec(ctx, `INSERT INTO users (id, name, username) VALUES ($1, 'Test', $2)`, f.userID, f.userID.String()); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO accounts (user_id, name, type) VALUES ($1, 'MariBank', 'ASSET') RETURNING id`, f.userID).Scan(&f.bankID); err != nil {
		t.Fatal(err)
	}
	return f
}

func date(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func money(s string) decimal.NullDecimal {
	return decimal.NullDecimal{Decimal: decimal.RequireFromString(s), Valid: true}
}

// statementFor builds a month with one money-out and one money-in row, each file distinct.
func (f *fixture) statementFor(month time.Month, opening, closing, out, in string) statement.Statement {
	f.fileSeed++
	start := date(2026, month, 1)
	return statement.Statement{
		UserID:         f.userID,
		AccountID:      f.bankID,
		Bank:           "MARIBANK",
		PeriodStart:    start,
		PeriodEnd:      start.AddDate(0, 1, -1),
		OpeningBalance: money(opening),
		ClosingBalance: money(closing),
		FileSHA256:     fmt.Sprintf("%064x", f.fileSeed),
		Rows: []statement.Row{
			{Date: start.AddDate(0, 0, 4), Description: "Transfer out", Amount: decimal.RequireFromString(out), MoneyOut: true},
			{Date: start.AddDate(0, 0, 9), Description: "Transfer in", Amount: decimal.RequireFromString(in)},
		},
	}
}

// The real MariBank statements: September ends where October starts.
func (f *fixture) september() statement.Statement {
	return f.statementFor(time.September, "8863.82", "24849.21", "106133.77", "122119.16")
}
func (f *fixture) october() statement.Statement {
	return f.statementFor(time.October, "24849.21", "1496.63", "149855.08", "126502.50")
}

func (f *fixture) importOK(st statement.Statement) statement.Result {
	f.t.Helper()
	result, err := f.service.Import(f.ctx, st)
	if err != nil {
		f.t.Fatal(err)
	}
	return result
}

func (f *fixture) balance() decimal.Decimal {
	f.t.Helper()
	var b decimal.Decimal
	if err := f.pool.QueryRow(f.ctx, `
		SELECT COALESCE(SUM(CASE WHEN type = 'DEBIT' THEN amount ELSE -amount END), 0) FROM entries WHERE account_id = $1
	`, f.bankID).Scan(&b); err != nil {
		f.t.Fatal(err)
	}
	return b
}

// openingBalances returns the signed amount of each opening-balance transaction on the bank account, by Manila date.
func (f *fixture) openingBalances() map[string]string {
	f.t.Helper()
	rows, err := f.pool.Query(f.ctx, `
		SELECT to_char(t.created_at AT TIME ZONE 'Asia/Manila', 'YYYY-MM-DD'),
		       CASE WHEN e.type = 'DEBIT' THEN e.amount ELSE -e.amount END
		FROM transactions t JOIN entries e ON e.transaction_id = t.id
		WHERE t.opening_balance AND e.account_id = $1
	`, f.bankID)
	if err != nil {
		f.t.Fatal(err)
	}
	defer rows.Close()
	got := map[string]string{}
	for rows.Next() {
		var day string
		var amount decimal.Decimal
		if err := rows.Scan(&day, &amount); err != nil {
			f.t.Fatal(err)
		}
		got[day] = amount.StringFixed(2)
	}
	return got
}

func (f *fixture) wantBalance(want string) {
	f.t.Helper()
	if got := f.balance(); !got.Equal(decimal.RequireFromString(want)) {
		f.t.Fatalf("balance = %s, want %s", got.StringFixed(2), want)
	}
}

func (f *fixture) wantOpeningBalances(want map[string]string) {
	f.t.Helper()
	got := f.openingBalances()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		f.t.Fatalf("opening balances = %v, want %v", got, want)
	}
}

func TestImportOldestFirst(t *testing.T) {
	f := setup(t)
	f.importOK(f.september())
	result := f.importOK(f.october())

	f.wantBalance("1496.63")
	f.wantOpeningBalances(map[string]string{"2026-09-01": "8863.82"})
	if len(result.Warnings) != 0 {
		t.Fatalf("warnings = %v", result.Warnings)
	}
}

func TestImportLaterMonthFirst(t *testing.T) {
	f := setup(t)
	f.importOK(f.october())
	f.wantOpeningBalances(map[string]string{"2026-10-01": "24849.21"})

	result := f.importOK(f.september())

	// One opening balance, moved to the earliest statement; October now follows from September.
	f.wantBalance("1496.63")
	f.wantOpeningBalances(map[string]string{"2026-09-01": "8863.82"})
	if len(result.Warnings) != 0 {
		t.Fatalf("warnings = %v", result.Warnings)
	}
}

func TestSameStatementTwiceIsRefusedAndChangesNothing(t *testing.T) {
	f := setup(t)
	october := f.october()
	f.importOK(october)

	if _, err := f.service.Import(f.ctx, october); !errors.Is(err, statement.ErrAlreadyImported) {
		t.Fatalf("err = %v, want ErrAlreadyImported", err)
	}
	// A re-exported file has a different hash but the same period: also refused.
	october.FileSHA256 = fmt.Sprintf("%064x", 999)
	if _, err := f.service.Import(f.ctx, october); !errors.Is(err, statement.ErrAlreadyImported) {
		t.Fatalf("err = %v, want ErrAlreadyImported", err)
	}
	f.wantBalance("1496.63")
}

func TestOverlappingStatementSkipsRowsAlreadyImported(t *testing.T) {
	f := setup(t)
	f.importOK(f.october())

	// Mid-October to mid-November: the 10 Oct row is already in; the 20 Oct and 5 Nov rows are new.
	f.fileSeed++
	overlap := statement.Statement{
		UserID: f.userID, AccountID: f.bankID, Bank: "MARIBANK",
		PeriodStart: date(2026, time.October, 10), PeriodEnd: date(2026, time.November, 9),
		FileSHA256: fmt.Sprintf("%064x", f.fileSeed),
		Rows: []statement.Row{
			{Date: date(2026, time.October, 10), Description: "Transfer in", Amount: decimal.RequireFromString("126502.50")},
			{Date: date(2026, time.October, 20), Description: "Coffee", Amount: decimal.RequireFromString("100"), MoneyOut: true},
			{Date: date(2026, time.November, 5), Description: "Coffee", Amount: decimal.RequireFromString("100"), MoneyOut: true},
		},
	}
	result := f.importOK(overlap)

	if result.Imported != 2 || len(result.Duplicates) != 1 {
		t.Fatalf("imported %d, duplicates %v; want 2 imported, 1 duplicate", result.Imported, result.Duplicates)
	}
	f.wantBalance("1296.63") // 1,496.63 − 100 − 100
}

func TestSameDayIdenticalRowsInOneStatementAllImport(t *testing.T) {
	f := setup(t)
	st := f.october()
	st.Rows = append(st.Rows, st.Rows[0]) // two identical transfers out on 5 Oct
	result := f.importOK(st)

	if result.Imported != 3 || len(result.Duplicates) != 0 {
		t.Fatalf("imported %d, duplicates %v; want 3 imported, none duplicate", result.Imported, result.Duplicates)
	}
}

func TestHandEnteredTransactionIsNotRecordedTwice(t *testing.T) {
	f := setup(t)
	var foodID uuid.UUID
	if err := f.pool.QueryRow(f.ctx, `INSERT INTO accounts (user_id, name, type) VALUES ($1, 'Food', 'EXPENSE') RETURNING id`, f.userID).Scan(&foodID); err != nil {
		t.Fatal(err)
	}
	// Entered by hand on 5 Oct (Manila) with its own description, before the statement arrives.
	at := time.Date(2026, time.October, 5, 0, 0, 0, 0, time.FixedZone("PHT", 8*60*60))
	if _, err := f.ledger.CreateTransaction(f.ctx, "Rent", []ledger.Entry{
		{AccountID: foodID, Amount: decimal.RequireFromString("149855.08"), Type: ledger.DEBIT},
		{AccountID: f.bankID, Amount: decimal.RequireFromString("149855.08"), Type: ledger.CREDIT},
	}, &at); err != nil {
		t.Fatal(err)
	}

	result := f.importOK(f.october())

	if result.Imported != 1 || len(result.Duplicates) != 1 {
		t.Fatalf("imported %d, duplicates %v; want 1 imported, 1 duplicate", result.Imported, result.Duplicates)
	}
	f.wantBalance("1496.63")
}

func TestMissingMonthIsReportedNotHidden(t *testing.T) {
	f := setup(t)
	// August ends at 8,000; September (opening 8,863.82) was never imported.
	f.importOK(f.statementFor(time.August, "5000.00", "8000.00", "1000.00", "4000.00"))
	result := f.importOK(f.october())

	if len(result.Warnings) == 0 {
		t.Fatal("expected a warning for the missing September statement")
	}
	// The ledger is the sum of what was imported; nothing was invented to cover the gap.
	f.wantBalance("-15352.58") // 8,000 − 149,855.08 + 126,502.50
	f.wantOpeningBalances(map[string]string{"2026-08-01": "5000.00"})
}

func TestStatementsCanBeReimportedAfterClearing(t *testing.T) {
	f := setup(t)
	october := f.october()
	f.importOK(october)

	if _, err := f.ledger.ClearTransactions(f.ctx, f.userID); err != nil {
		t.Fatal(err)
	}
	f.wantBalance("0")

	f.importOK(october)
	f.wantBalance("1496.63")
}

// A bank statement into a credit card account (or the reverse) is refused: the signs would mean the opposite.
func TestStatementMustMatchTheAccountKind(t *testing.T) {
	f := setup(t)
	var cardID uuid.UUID
	if err := f.pool.QueryRow(f.ctx, `INSERT INTO accounts (user_id, name, type) VALUES ($1, 'Card', 'CREDIT_CARD') RETURNING id`, f.userID).Scan(&cardID); err != nil {
		t.Fatal(err)
	}
	st := f.october()
	st.AccountID = cardID
	if _, err := f.service.Import(f.ctx, st); !errors.Is(err, statement.ErrInvalidStatement) {
		t.Fatalf("err = %v, want ErrInvalidStatement", err)
	}
}

func (f *fixture) accountNumber(id uuid.UUID) string {
	f.t.Helper()
	var number *string
	if err := f.pool.QueryRow(f.ctx, `SELECT account_number FROM accounts WHERE id = $1`, id).Scan(&number); err != nil {
		f.t.Fatal(err)
	}
	if number == nil {
		return ""
	}
	return *number
}

func TestAccountIsFoundByNumberOrCreated(t *testing.T) {
	f := setup(t)
	october := f.october()
	october.AccountID, october.AccountNumber, october.AccountName = uuid.Nil, "12345678901", "MariBank ••8901"
	first := f.importOK(october)

	if first.AccountID == f.bankID || first.AccountName != "MariBank ••8901" {
		t.Fatalf("got account %s %q, want a new account named MariBank ••8901", first.AccountID, first.AccountName)
	}

	september := f.september()
	september.AccountID, september.AccountNumber, september.AccountName = uuid.Nil, "12345678901", "ignored"
	second := f.importOK(september)

	if second.AccountID != first.AccountID {
		t.Fatalf("September went into %s, want the account October created (%s)", second.AccountID, first.AccountID)
	}
	if len(second.Warnings) != 0 {
		t.Fatalf("warnings = %v", second.Warnings)
	}
}

func TestChosenAccountLearnsTheNumber(t *testing.T) {
	f := setup(t)
	october := f.october()
	october.AccountNumber = "12345678901"
	f.importOK(october)

	if got := f.accountNumber(f.bankID); got != "12345678901" {
		t.Fatalf("account number = %q, want it saved on the chosen account", got)
	}

	// From now on the number alone finds it.
	september := f.september()
	september.AccountID, september.AccountNumber, september.AccountName = uuid.Nil, "12345678901", "unused"
	if result := f.importOK(september); result.AccountID != f.bankID {
		t.Fatalf("September went into %s, want %s", result.AccountID, f.bankID)
	}
}

func TestStatementForAnotherAccountNumberIsRefused(t *testing.T) {
	f := setup(t)
	october := f.october()
	october.AccountNumber = "12345678901"
	f.importOK(october)

	september := f.september()
	september.AccountNumber = "99999999999"
	if _, err := f.service.Import(f.ctx, september); !errors.Is(err, statement.ErrInvalidStatement) {
		t.Fatalf("err = %v, want ErrInvalidStatement", err)
	}
	f.wantBalance("1496.63")
}

func TestOtherUsersAccountIsRefused(t *testing.T) {
	f := setup(t)
	st := f.october()
	st.UserID = uuid.New()
	if _, err := f.service.Import(f.ctx, st); !errors.Is(err, statement.ErrInvalidStatement) {
		t.Fatalf("err = %v, want ErrInvalidStatement", err)
	}
}

func TestTransfersGoToTheirOwnAccounts(t *testing.T) {
	f := setup(t)
	if _, err := f.pool.Exec(f.ctx, `
		INSERT INTO accounts (user_id, name, type) VALUES ($1, 'Salary Income', 'INCOME'), ($1, 'Uncategorized Expenses', 'EXPENSE')
	`, f.userID); err != nil {
		t.Fatal(err)
	}
	f.importOK(f.october())

	rows, err := f.pool.Query(f.ctx, `
		SELECT a.name FROM entries e JOIN accounts a ON a.id = e.account_id
		WHERE a.user_id = $1 AND a.type IN ('INCOME', 'EXPENSE') ORDER BY a.name
	`, f.userID)
	if err != nil {
		t.Fatal(err)
	}
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(names) != "[Incoming Transfers Outgoing Transfers]" {
		t.Fatalf("income/expense entries went to %v", names)
	}
}

// payroll adds a second bank account and a statement for it with the given rows (no printed balances).
func (f *fixture) payroll(rows ...statement.Row) statement.Statement {
	f.t.Helper()
	var id uuid.UUID
	err := f.pool.QueryRow(f.ctx, `SELECT id FROM accounts WHERE user_id = $1 AND name = 'UnionBank'`, f.userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = f.pool.QueryRow(f.ctx, `INSERT INTO accounts (user_id, name, type) VALUES ($1, 'UnionBank', 'ASSET') RETURNING id`, f.userID).Scan(&id)
	}
	if err != nil {
		f.t.Fatal(err)
	}
	f.fileSeed++
	return statement.Statement{
		UserID: f.userID, AccountID: id, Bank: "UNIONBANK",
		PeriodStart: date(2026, time.October, 1), PeriodEnd: date(2026, time.October, 31),
		FileSHA256: fmt.Sprintf("%064x", f.fileSeed),
		Rows:       rows,
	}
}

// entriesOn counts entries on the user's account with that name.
func (f *fixture) entriesOn(name string) int {
	f.t.Helper()
	var n int
	if err := f.pool.QueryRow(f.ctx, `
		SELECT count(*) FROM entries e JOIN accounts a ON a.id = e.account_id WHERE a.user_id = $1 AND a.name = $2
	`, f.userID, name).Scan(&n); err != nil {
		f.t.Fatal(err)
	}
	return n
}

// October's "Transfer in" of 126,502.50 on 10 Oct came from the payroll account.
var payrollOut = statement.Row{Date: date(2026, time.October, 10), Description: "Sent to SEA", Amount: decimal.RequireFromString("126502.50"), MoneyOut: true}

func TestTransferJoinedWhenOtherBankIsImportedSecond(t *testing.T) {
	f := setup(t)
	f.importOK(f.october())
	result := f.importOK(f.payroll(payrollOut))

	if len(result.Transfers) != 1 || result.Imported != 0 {
		t.Fatalf("transfers %v, imported %d; want 1 transfer, nothing new", result.Transfers, result.Imported)
	}
	// Neither side is income or spending any more.
	if got := f.entriesOn("Incoming Transfers"); got != 0 {
		t.Fatalf("Incoming Transfers has %d entries, want 0", got)
	}
	f.wantBalance("1496.63")
}

func TestTransferJoinedWhenThisBankIsImportedSecond(t *testing.T) {
	f := setup(t)
	f.importOK(f.payroll(payrollOut))
	result := f.importOK(f.october())

	if len(result.Transfers) != 1 || result.Imported != 1 {
		t.Fatalf("transfers %v, imported %d; want 1 transfer and the transfer out recorded", result.Transfers, result.Imported)
	}
	if got := f.entriesOn("Outgoing Transfers"); got != 1 { // only October's own money out
		t.Fatalf("Outgoing Transfers has %d entries, want 1", got)
	}
	f.wantBalance("1496.63")
}

func TestTransferDatedNextDayByTheOtherBankStillJoins(t *testing.T) {
	f := setup(t)
	f.importOK(f.october())
	late := payrollOut
	late.Date = date(2026, time.October, 11)
	if result := f.importOK(f.payroll(late)); len(result.Transfers) != 1 {
		t.Fatalf("transfers = %v, want the 11 Oct row joined with 10 Oct", result.Transfers)
	}
}

func TestJoinedTransferIsADuplicateInAnOverlappingStatement(t *testing.T) {
	f := setup(t)
	f.importOK(f.october())
	late := payrollOut
	late.Date = date(2026, time.October, 11) // the payroll bank's own date; the transfer carries 10 Oct
	f.importOK(f.payroll(late))

	again := f.payroll(late)
	again.PeriodStart, again.PeriodEnd = date(2026, time.October, 5), date(2026, time.November, 4)
	result := f.importOK(again)
	if len(result.Duplicates) != 1 || len(result.Transfers) != 0 || result.Imported != 0 {
		t.Fatalf("duplicates %v, transfers %v, imported %d; want it recognised as already recorded",
			result.Duplicates, result.Transfers, result.Imported)
	}
}

func TestEquallyCloseCandidatesAreNotGuessed(t *testing.T) {
	f := setup(t)
	st := f.october()
	// Two ₱500 transfers in, on 9 and 11 Oct: a ₱500 out on 10 Oct could be either.
	st.Rows = append(st.Rows,
		statement.Row{Date: date(2026, time.October, 9), Description: "In", Amount: decimal.RequireFromString("500")},
		statement.Row{Date: date(2026, time.October, 11), Description: "In", Amount: decimal.RequireFromString("500")})
	st.ClosingBalance = decimal.NullDecimal{}
	f.importOK(st)

	result := f.importOK(f.payroll(statement.Row{Date: date(2026, time.October, 10), Description: "Out", Amount: decimal.RequireFromString("500"), MoneyOut: true}))
	if len(result.Transfers) != 0 || result.Imported != 1 {
		t.Fatalf("transfers %v, imported %d; want it recorded as spending, not guessed", result.Transfers, result.Imported)
	}
}

func TestCategorisedRowGoesToItsAccountAndIsNeverATransfer(t *testing.T) {
	f := setup(t)
	f.importOK(f.october())
	// Same amount and day as October's "Transfer in", but it's payroll.
	salary := payrollOut
	salary.MoneyOut, salary.Category = false, "Salary Income"
	result := f.importOK(f.payroll(salary))

	if len(result.Transfers) != 0 || result.Imported != 1 {
		t.Fatalf("transfers %v, imported %d; want salary recorded, not joined", result.Transfers, result.Imported)
	}
	if got := f.entriesOn("Salary Income"); got != 1 {
		t.Fatalf("Salary Income has %d entries, want 1", got)
	}
}

// A credit card statement: balances are owed (negative), charges are money out, a payment is money in,
// and the payment joins the bank row that paid it.
func TestCreditCardStatement(t *testing.T) {
	f := setup(t)
	// The bank paid the previous bill on 25 Sep.
	bank := f.payroll(statement.Row{Date: date(2026, time.September, 25), Description: "BILLS PAYMENT", Amount: decimal.RequireFromString("14485.15"), MoneyOut: true})
	bank.PeriodStart, bank.PeriodEnd = date(2026, time.September, 1), date(2026, time.September, 30)
	f.importOK(bank)

	f.fileSeed++
	card := statement.Statement{
		UserID: f.userID, AccountNumber: "1135", AccountName: "Test Card", AccountType: "CREDIT_CARD", Bank: "UNIONBANK_CARD",
		PeriodStart: date(2026, time.September, 24), PeriodEnd: date(2026, time.October, 23),
		OpeningBalance: money("-14485.15"), ClosingBalance: money("-21247.65"),
		FileSHA256: fmt.Sprintf("%064x", f.fileSeed),
		Rows: []statement.Row{
			{Date: date(2026, time.September, 25), Description: "Payment via UB Online", Amount: decimal.RequireFromString("14485.15")},
			{Date: date(2026, time.September, 28), Description: "Rent", Amount: decimal.RequireFromString("17000"), MoneyOut: true},
			{Date: date(2026, time.October, 23), Description: "Instalment Interest", Amount: decimal.RequireFromString("4247.65"), MoneyOut: true, Category: "Loan Interest"},
		},
	}
	result := f.importOK(card)

	if len(result.Transfers) != 1 || result.Imported != 2 || len(result.Warnings) != 0 {
		t.Fatalf("transfers %v, imported %d, warnings %v; want the payment joined, 2 charges, balances matching",
			result.Transfers, result.Imported, result.Warnings)
	}
	var owed decimal.Decimal
	if err := f.pool.QueryRow(f.ctx, `SELECT SUM(CASE WHEN type = 'DEBIT' THEN amount ELSE -amount END) FROM entries WHERE account_id = $1`, result.AccountID).Scan(&owed); err != nil {
		t.Fatal(err)
	}
	if !owed.Equal(decimal.RequireFromString("-21247.65")) {
		t.Fatalf("card balance = %s, want -21247.65 (the total due)", owed)
	}

	// A bank statement can't go into the card.
	wrong := f.october()
	wrong.AccountID, wrong.AccountNumber = result.AccountID, ""
	if _, err := f.service.Import(f.ctx, wrong); !errors.Is(err, statement.ErrInvalidStatement) {
		t.Fatalf("err = %v, want ErrInvalidStatement", err)
	}
}
