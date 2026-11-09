package ledger

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo  *LedgerRepository
	redis *redis.Client
}

func NewService(repo *LedgerRepository, redis *redis.Client) *Service {
	return &Service{repo: repo, redis: redis}
}

func (s *Service) CreateTransaction(
	ctx context.Context,
	description string,
	entries []Entry,
	occurredAt *time.Time, // nil = now
) (uuid.UUID, error) {

	if err := validateEntries(entries); err != nil {
		return uuid.Nil, err
	}

	userID, err := s.repo.GetUserIDFromAccount(ctx, entries[0].AccountID)
	if err != nil {
		return uuid.Nil, err
	}

	// The transaction is stored against one user, so an entry on another user's account
	// would change their balance without appearing in their transaction list.
	for _, entry := range entries[1:] {
		owner, err := s.repo.GetUserIDFromAccount(ctx, entry.AccountID)
		if err != nil {
			return uuid.Nil, err
		}
		if owner != userID {
			return uuid.Nil, fmt.Errorf("%w: all accounts must belong to the same user", ErrInvalidTransaction)
		}
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	// If Commit succeeds, Rollback does nothing.
	defer tx.Rollback(ctx)

	transactionID, err := s.repo.CreateTransactionTx(ctx, tx, userID, description, occurredAt)
	if err != nil {
		return uuid.Nil, err
	}

	accountIDs := make([]uuid.UUID, 0, len(entries))

	for _, entry := range entries {
		if _, err := s.repo.InsertEntryTx(ctx, tx, transactionID, entry.AccountID, entry.Amount, entry.Type); err != nil {
			return uuid.Nil, err
		}

		accountIDs = append(accountIDs, entry.AccountID)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}

	// Invalidate caches only AFTER successful commit.
	cache.InvalidateTransaction(ctx, s.redis, userID, accountIDs...)

	return transactionID, nil
}

// ClearTransactions deletes all of the user's transactions and keeps their accounts.
func (s *Service) ClearTransactions(ctx context.Context, userID uuid.UUID) (int64, error) {
	accounts, err := s.repo.GetUserAccounts(ctx, userID)
	if err != nil {
		return 0, err
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	deleted, err := s.repo.DeleteUserTransactionsTx(ctx, tx, userID)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	// Every balance, summary and ledger of the user's accounts is now stale.
	accountIDs := make([]uuid.UUID, 0, len(accounts))
	for _, account := range accounts {
		if account.Type == string(CREDIT_CARD) {
			cache.InvalidateCreditCard(ctx, s.redis, account.ID, userID)
		}
		accountIDs = append(accountIDs, account.ID)
	}
	cache.InvalidateTransaction(ctx, s.redis, userID, accountIDs...)

	return deleted, nil
}

// UpdateTransaction changes one of the user's transactions: its description, date (nil keeps it) and each
// entry's account, amount and type. Every current entry must be given, by id; none can be added or removed.
func (s *Service) UpdateTransaction(
	ctx context.Context,
	userID uuid.UUID,
	transactionID uuid.UUID,
	description string,
	occurredAt *time.Time,
	entries []Entry,
) error {
	if err := validateEntries(entries); err != nil {
		return err
	}
	current, err := s.ownTransaction(ctx, userID, transactionID)
	if err != nil {
		return err
	}
	if len(entries) != len(current.Entries) {
		return fmt.Errorf("%w: entries can't be added or removed", ErrInvalidTransaction)
	}

	remaining := map[uuid.UUID]bool{}
	touched := make([]uuid.UUID, 0, 2*len(entries))
	for _, e := range current.Entries {
		remaining[e.ID] = true
		touched = append(touched, e.AccountID)
	}
	for _, e := range entries {
		if !remaining[e.ID] {
			return fmt.Errorf("%w: entry %s isn't part of this transaction", ErrInvalidTransaction, e.ID)
		}
		delete(remaining, e.ID) // each entry once
		owner, err := s.repo.GetUserIDFromAccount(ctx, e.AccountID)
		if err != nil {
			return err
		}
		if owner != userID {
			return fmt.Errorf("%w: all accounts must belong to the same user", ErrInvalidTransaction)
		}
		touched = append(touched, e.AccountID)
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := s.repo.UpdateTransactionTx(ctx, tx, transactionID, description, occurredAt, entries); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.invalidate(ctx, userID, touched)
	return nil
}

// DeleteTransaction deletes one of the user's transactions.
func (s *Service) DeleteTransaction(ctx context.Context, userID, transactionID uuid.UUID) error {
	current, err := s.ownTransaction(ctx, userID, transactionID)
	if err != nil {
		return err
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := s.repo.DeleteTransactionTx(ctx, tx, transactionID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	touched := make([]uuid.UUID, 0, len(current.Entries))
	for _, e := range current.Entries {
		touched = append(touched, e.AccountID)
	}
	s.invalidate(ctx, userID, touched)
	return nil
}

// ownTransaction loads a transaction, or pgx.ErrNoRows (not found) when it isn't one of the user's.
func (s *Service) ownTransaction(ctx context.Context, userID, transactionID uuid.UUID) (*TransactionDetail, error) {
	t, err := s.repo.GetTransaction(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	// Every entry is on one user's accounts (CreateTransaction checks), so the first one tells whose it is.
	if len(t.Entries) == 0 {
		return nil, pgx.ErrNoRows
	}
	owner, err := s.repo.GetUserIDFromAccount(ctx, t.Entries[0].AccountID)
	if err != nil {
		return nil, err
	}
	if owner != userID {
		return nil, pgx.ErrNoRows
	}
	return t, nil
}

// invalidate drops the cached balances, ledgers and summaries of the accounts a change touched.
func (s *Service) invalidate(ctx context.Context, userID uuid.UUID, accountIDs []uuid.UUID) {
	for _, id := range accountIDs {
		if _, accountType, _, err := s.repo.GetAccount(ctx, id); err == nil && accountType == CREDIT_CARD {
			cache.InvalidateCreditCard(ctx, s.redis, id, userID)
		}
	}
	cache.InvalidateTransaction(ctx, s.redis, userID, accountIDs...)
}

func (s *Service) GetUserAccounts(ctx context.Context, userID uuid.UUID) ([]Account, error) {
	return cache.GetOrLoad(ctx, s.redis, cache.UserAccountsKey(userID), 5*time.Minute, func() ([]Account, error) {
		return s.repo.GetUserAccounts(ctx, userID)
	})
}

func (s *Service) GetBalance(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	return cache.GetOrLoad(ctx, s.redis, cache.AccountBalanceKey(accountID), 5*time.Minute, func() (decimal.Decimal, error) {
		return s.repo.GetBalance(ctx, accountID)
	})
}

func (s *Service) GetAccountSummary(ctx context.Context, accountID uuid.UUID) (*AccountSummary, error) {
	return cache.GetOrLoad(ctx, s.redis, cache.AccountSummaryKey(accountID), 10*time.Minute, func() (*AccountSummary, error) {
		name, accountType, createdAt, err := s.repo.GetAccount(ctx, accountID)
		if err != nil {
			return nil, err
		}

		balance, err := s.repo.GetBalance(ctx, accountID)
		if err != nil {
			return nil, err
		}

		return &AccountSummary{
			AccountID: accountID,
			Name:      name,
			Type:      accountType,
			Balance:   balance,
			CreatedAt: createdAt,
		}, nil
	})
}

func (s *Service) GetLedgerEntries(ctx context.Context, accountID uuid.UUID) ([]Entry, error) {
	return cache.GetOrLoad(ctx, s.redis, cache.AccountLedgerEntriesKey(accountID), 2*time.Minute, func() ([]Entry, error) {
		return s.repo.GetLedgerEntries(ctx, accountID)
	})
}

func (s *Service) GetUserTotalCredit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	return cache.GetOrLoad(ctx, s.redis, cache.UserTotalCreditKey(userID), 5*time.Minute, func() (decimal.Decimal, error) {
		return s.repo.GetUserTotalCredit(ctx, userID)
	})
}

func (s *Service) GetUserTotalDebit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	return cache.GetOrLoad(ctx, s.redis, cache.UserTotalDebitKey(userID), 5*time.Minute, func() (decimal.Decimal, error) {
		return s.repo.GetUserTotalDebit(ctx, userID)
	})
}

func (s *Service) GetUserNetWorth(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	return cache.GetOrLoad(ctx, s.redis, cache.UserNetWorthKey(userID), 5*time.Minute, func() (decimal.Decimal, error) {
		return s.repo.GetUserNetWorth(ctx, userID)
	})
}

func (s *Service) GetUserRecentTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]RecentTransaction, error) {
	return s.repo.GetUserRecentTransactions(ctx, userID, limit)
}

func (s *Service) GetMoneyFlows(ctx context.Context, userID uuid.UUID, from, to *time.Time) ([]MoneyFlow, error) {
	return s.repo.GetMoneyFlows(ctx, userID, from, to)
}

func (s *Service) GetTransaction(ctx context.Context, transactionID uuid.UUID) (*TransactionDetail, error) {
	return s.repo.GetTransaction(ctx, transactionID)
}

func (s *Service) GetUserTransactions(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]TransactionDetail, int, error) {
	return s.repo.GetUserTransactions(ctx, userID, limit, offset)
}
