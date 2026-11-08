package ledger

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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

	transactionID, err := s.repo.CreateTransactionTx(ctx, tx, userID, description)
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

func (s *Service) GetTransaction(ctx context.Context, transactionID uuid.UUID) (*TransactionDetail, error) {
	return s.repo.GetTransaction(ctx, transactionID)
}

func (s *Service) GetUserTransactions(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]TransactionDetail, int, error) {
	return s.repo.GetUserTransactions(ctx, userID, limit, offset)
}
