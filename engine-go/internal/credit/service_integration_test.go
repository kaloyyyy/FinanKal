package credit_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kaloy/finankal/engine-go/internal/credit"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

// These tests write rows, so point TEST_DATABASE_URL at a throwaway database loaded from
// db/schema.sql, never at real data. TEST_REDIS_ADDR is optional and turns caching on.

var purchaseDate = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

type fixture struct {
	pool   *pgxpool.Pool
	credit *credit.Service
	ledger *ledger.Service
	userID uuid.UUID
	cardID uuid.UUID
	bankID uuid.UUID
	ctx    context.Context
	t      *testing.T
}

func amount(value string) decimal.Decimal { return decimal.RequireFromString(value) }

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

	var rdb *redis.Client
	if addr := os.Getenv("TEST_REDIS_ADDR"); addr != "" {
		rdb = redis.NewClient(&redis.Options{Addr: addr})
		t.Cleanup(func() { rdb.Close() })
	}

	ledgerRepo := ledger.NewLedgerRepository(pool)
	f := &fixture{
		pool:   pool,
		ctx:    ctx,
		t:      t,
		ledger: ledger.NewService(ledgerRepo, rdb),
		credit: credit.NewService(
			credit.NewCardRepository(pool),
			ledgerRepo,
			credit.NewStatementRepository(pool),
			credit.NewEntryRepository(pool),
			rdb,
		),
		userID: uuid.New(),
	}

	f.insertUser(f.userID)
	f.bankID = f.insertAccount(f.userID, "Bank", "ASSET")
	f.cardID = f.createCard("500")

	return f
}

func (f *fixture) insertUser(id uuid.UUID) {
	f.t.Helper()
	_, err := f.pool.Exec(f.ctx, `INSERT INTO users (id, name, username) VALUES ($1, 'Test', $2)`, id, id.String())
	if err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) insertAccount(userID uuid.UUID, name, accountType string) uuid.UUID {
	f.t.Helper()
	var id uuid.UUID
	err := f.pool.QueryRow(f.ctx, `INSERT INTO accounts (user_id, name, type) VALUES ($1,$2,$3) RETURNING id`, userID, name, accountType).Scan(&id)
	if err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *fixture) createCard(limit string) uuid.UUID {
	f.t.Helper()
	account := f.insertAccount(f.userID, "Card", "CREDIT_CARD")
	id, err := f.credit.CreateCreditCard(f.ctx, account, "", amount(limit), 15, 20)
	if err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *fixture) purchase(cardID uuid.UUID, value string, when time.Time) error {
	_, err := f.credit.RecordCreditCardTransaction(f.ctx, credit.CreditCardTransactionRequest{
		CardID: cardID, Amount: amount(value), Description: "purchase", PurchaseDate: when,
	})
	return err
}

func (f *fixture) pay(cardID, statementID, fromAccount uuid.UUID, value string) error {
	_, err := f.credit.PayCreditCardStatement(f.ctx, credit.CreditCardPaymentRequest{
		CardID: cardID, StatementID: statementID, PaymentAccountID: fromAccount, Amount: amount(value), Description: "payment",
	})
	return err
}

func (f *fixture) statementID(cardID uuid.UUID) uuid.UUID {
	f.t.Helper()
	var id uuid.UUID
	err := f.pool.QueryRow(f.ctx, `SELECT id FROM credit_card_statements WHERE credit_card_id = $1`, cardID).Scan(&id)
	if err != nil {
		f.t.Fatal(err)
	}
	return id
}

// warmPool opens n connections up front. Opening one takes far longer than a whole payment,
// so on a cold pool the first goroutine finishes before the rest connect and nothing overlaps.
func (f *fixture) warmPool(n int) {
	f.t.Helper()
	conns := make([]*pgxpool.Conn, n)
	for i := range conns {
		conn, err := f.pool.Acquire(f.ctx)
		if err != nil {
			f.t.Fatal(err)
		}
		conns[i] = conn
	}
	for _, conn := range conns {
		conn.Release()
	}
}

func wantErr(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCreditLimitIsEnforced(t *testing.T) {
	f := setup(t)

	if err := f.purchase(f.cardID, "400", purchaseDate); err != nil {
		t.Fatal(err)
	}
	wantErr(t, f.purchase(f.cardID, "100.01", purchaseDate), credit.ErrCreditLimitExceeded)

	if err := f.purchase(f.cardID, "100", purchaseDate); err != nil {
		t.Fatalf("spending exactly up to the limit: %v", err)
	}
}

func TestConcurrentPurchasesCannotExceedLimit(t *testing.T) {
	f := setup(t)

	f.warmPool(20)

	var wg sync.WaitGroup
	results := make([]error, 20)
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = f.purchase(f.cardID, "100", purchaseDate)
		}()
	}
	wg.Wait()

	succeeded := 0
	for _, err := range results {
		switch {
		case err == nil:
			succeeded++
		case !errors.Is(err, credit.ErrCreditLimitExceeded):
			t.Fatal(err)
		}
	}
	if succeeded != 5 {
		t.Errorf("%d purchases of 100 succeeded on a limit of 500, want 5", succeeded)
	}

	var statements int
	if err := f.pool.QueryRow(f.ctx, `SELECT count(*) FROM credit_card_statements WHERE credit_card_id = $1`, f.cardID).Scan(&statements); err != nil {
		t.Fatal(err)
	}
	if statements != 1 {
		t.Errorf("got %d statements for one cycle, want 1", statements)
	}
}

func TestPaymentRules(t *testing.T) {
	f := setup(t)

	if err := f.purchase(f.cardID, "300", purchaseDate); err != nil {
		t.Fatal(err)
	}
	statementID := f.statementID(f.cardID)

	otherCard := f.createCard("500")
	otherUser := uuid.New()
	f.insertUser(otherUser)

	wantErr(t, f.pay(otherCard, statementID, f.bankID, "10"), credit.ErrStatementCardMismatch)
	wantErr(t, f.pay(f.cardID, statementID, f.insertAccount(otherUser, "Theirs", "ASSET"), "10"), credit.ErrInvalidPaymentAccount)
	wantErr(t, f.pay(f.cardID, statementID, f.cardID, "10"), credit.ErrInvalidPaymentAccount)
	wantErr(t, f.pay(f.cardID, statementID, f.bankID, "300.01"), credit.ErrPaymentTooLarge)
	wantErr(t, f.pay(f.cardID, statementID, f.bankID, "-5"), credit.ErrInvalidAmount)
	wantErr(t, f.pay(f.cardID, statementID, f.bankID, "1.005"), credit.ErrInvalidAmount)

	if err := f.pay(f.cardID, statementID, f.bankID, "100"); err != nil {
		t.Fatal(err)
	}
	statement, err := f.credit.GetStatement(f.ctx, statementID)
	if err != nil {
		t.Fatal(err)
	}
	if !statement.TotalAmount.Equal(amount("200")) || statement.Status != credit.StatementOpen {
		t.Errorf("after partial payment: %s %s", statement.TotalAmount, statement.Status)
	}

	if err := f.pay(f.cardID, statementID, f.bankID, "200"); err != nil {
		t.Fatal(err)
	}
	wantErr(t, f.pay(f.cardID, statementID, f.bankID, "1"), credit.ErrStatementAlreadyPaid)

	balance, err := f.ledger.GetBalance(f.ctx, f.cardID)
	if err != nil {
		t.Fatal(err)
	}
	if !balance.IsZero() {
		t.Errorf("card balance after paying in full: %s", balance)
	}
}

func TestConcurrentPaymentsCannotOverpay(t *testing.T) {
	f := setup(t)

	if err := f.purchase(f.cardID, "100", purchaseDate); err != nil {
		t.Fatal(err)
	}
	statementID := f.statementID(f.cardID)

	f.warmPool(10)

	var wg sync.WaitGroup
	results := make([]error, 10)
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = f.pay(f.cardID, statementID, f.bankID, "100")
		}()
	}
	wg.Wait()

	succeeded := 0
	for _, err := range results {
		if err == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Errorf("%d full payments succeeded, want 1", succeeded)
	}

	balance, err := f.ledger.GetBalance(f.ctx, f.bankID)
	if err != nil {
		t.Fatal(err)
	}
	if !balance.Equal(amount("-100")) {
		t.Errorf("bank balance %s, want -100", balance)
	}
}

func TestPurchaseAfterFullPaymentReopensStatement(t *testing.T) {
	f := setup(t)

	if err := f.purchase(f.cardID, "100", purchaseDate); err != nil {
		t.Fatal(err)
	}
	statementID := f.statementID(f.cardID)
	if err := f.pay(f.cardID, statementID, f.bankID, "100"); err != nil {
		t.Fatal(err)
	}

	if err := f.purchase(f.cardID, "30", purchaseDate); err != nil {
		t.Fatal(err)
	}

	statement, err := f.credit.GetStatement(f.ctx, statementID)
	if err != nil {
		t.Fatal(err)
	}
	if statement.Status != credit.StatementOpen || !statement.TotalAmount.Equal(amount("30")) {
		t.Fatalf("got %s %s, want OPEN 30", statement.Status, statement.TotalAmount)
	}
	if err := f.pay(f.cardID, statementID, f.bankID, "30"); err != nil {
		t.Fatal(err)
	}
}

func TestBalancesAreFreshAfterCardActivity(t *testing.T) {
	f := setup(t)

	cardBalance := func() string {
		t.Helper()
		accounts, err := f.ledger.GetUserAccounts(f.ctx, f.userID)
		if err != nil {
			t.Fatal(err)
		}
		for _, account := range accounts {
			if account.ID == f.cardID {
				return account.Balance.String()
			}
		}
		t.Fatal("card account missing")
		return ""
	}

	cardBalance() // fills the cache

	if err := f.purchase(f.cardID, "100", purchaseDate); err != nil {
		t.Fatal(err)
	}
	if got := cardBalance(); got != "-100" {
		t.Fatalf("after purchase: %s", got)
	}

	if err := f.pay(f.cardID, f.statementID(f.cardID), f.bankID, "40"); err != nil {
		t.Fatal(err)
	}
	if got := cardBalance(); got != "-60" {
		t.Errorf("after payment: %s", got)
	}
}

func TestCreateCreditCardLeavesNoOrphansOnFailure(t *testing.T) {
	f := setup(t)

	countAccounts := func() int {
		t.Helper()
		var n int
		if err := f.pool.QueryRow(f.ctx, `SELECT count(*) FROM accounts WHERE user_id = $1`, f.userID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	before := countAccounts()

	// The account already has a card, so inserting a second one violates the primary key.
	if _, err := f.credit.CreateCreditCard(f.ctx, f.cardID, "", amount("100"), 15, 20); err == nil {
		t.Fatal("want an error for a duplicate card")
	}

	if after := countAccounts(); after != before {
		t.Errorf("accounts went from %d to %d after a failed create", before, after)
	}
}

func TestLedgerRejectsCrossUserTransaction(t *testing.T) {
	f := setup(t)

	otherUser := uuid.New()
	f.insertUser(otherUser)

	_, err := f.ledger.CreateTransaction(f.ctx, "cross user", []ledger.Entry{
		{AccountID: f.bankID, Amount: amount("10"), Type: ledger.DEBIT},
		{AccountID: f.insertAccount(otherUser, "Theirs", "ASSET"), Amount: amount("10"), Type: ledger.CREDIT},
	})
	wantErr(t, err, ledger.ErrInvalidTransaction)
}

func TestMissingStatementIsNoRows(t *testing.T) {
	f := setup(t)

	_, err := f.credit.GetStatement(f.ctx, uuid.New())
	wantErr(t, err, pgx.ErrNoRows)
}
