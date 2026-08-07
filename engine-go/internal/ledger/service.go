package ledger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/shopspring/decimal"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo  *LedgerRepository
	redis *redis.Client
}

func NewService(repo *LedgerRepository, redis *redis.Client) *Service {
	return &Service{repo: repo, redis: redis}
}

func (s *Service) CreateTransaction(ctx context.Context, description string, entries []Entry) (uuid.UUID, error) {
	if err := validateEntries(entries); err != nil {
		return uuid.UUID{}, err
	}

	// prevent panic if caller passed empty entries slice
	if len(entries) == 0 {
		return uuid.UUID{}, errors.New("create transaction: no entries provided")
	}

	// Get user_id from the first account
	userID, err := s.repo.GetUserIDFromAccount(ctx, entries[0].AccountID)
	if err != nil {
		return uuid.UUID{}, err
	}

	txID, err := s.repo.CreateTransaction(ctx, userID, description)
	if err != nil {
		return uuid.UUID{}, err
	}

	for _, e := range entries {
		_, err := s.repo.InsertEntry(ctx, txID, e.AccountID, e.Amount, e.Type)
		if err != nil {
			return uuid.UUID{}, err
		}
	}

	// Invalidate account-level caches using shared helper (safe if redis == nil)
	var accountIDs []uuid.UUID
	for _, e := range entries {
		accountIDs = append(accountIDs, e.AccountID)
	}

	_ = cache.InvalidateTransaction(
		ctx,
		s.redis,
		userID,
		accountIDs...,
	)

	return txID, nil
}

func (s *Service) GetBalance(
	ctx context.Context,
	accountID uuid.UUID,
) (decimal.Decimal, error) {

	key := cache.AccountBalanceKey(accountID)

	value, err := cache.Get[string](
		ctx,
		s.redis,
		key,
	)

	if err == nil && value != nil {

		return decimal.NewFromString(*value)
	}

	balance, err := s.repo.GetBalance(
		ctx,
		accountID,
	)

	if err != nil {
		return decimal.Zero, err
	}

	_ = cache.Set(
		ctx,
		s.redis,
		key,
		balance.String(),
		5*time.Minute,
	)

	return balance, nil
}

func (s *Service) GetAccountSummary(
	ctx context.Context,
	accountID uuid.UUID,
) (*AccountSummary, error) {

	key := cache.AccountSummaryKey(accountID)

	summary, err := cache.Get[AccountSummary](
		ctx,
		s.redis,
		key,
	)

	if err == nil && summary != nil {
		return summary, nil
	}

	name, accountType, createdAt, err :=
		s.repo.GetAccount(
			ctx,
			accountID,
		)

	if err != nil {
		return nil, err
	}

	balance, err :=
		s.repo.GetBalance(
			ctx,
			accountID,
		)

	if err != nil {
		return nil, err
	}

	result := &AccountSummary{
		AccountID: accountID,
		Name:      name,
		Type:      accountType,
		Balance:   balance,
		CreatedAt: createdAt,
	}

	_ = cache.Set(
		ctx,
		s.redis,
		key,
		result,
		10*time.Minute,
	)

	return result, nil
}

func (s *Service) CreateAccount(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	accountType AccountType,
) (uuid.UUID, error) {

	accountID, err := s.repo.CreateAccount(
		ctx,
		userID,
		name,
		accountType,
	)

	if err != nil {
		return uuid.Nil, err
	}

	// invalidate user caches
	_ = cache.InvalidateUser(
		ctx,
		s.redis,
		userID,
	)

	return accountID, nil
}

func (s *Service) GetLedgerEntries(
	ctx context.Context,
	accountID uuid.UUID,
) ([]Entry, error) {

	key := cache.AccountLedgerEntriesKey(accountID)

	entries, err :=
		cache.Get[[]Entry](
			ctx,
			s.redis,
			key,
		)

	if err == nil && entries != nil {
		return *entries, nil
	}

	entriesDB, err :=
		s.repo.GetLedgerEntries(
			ctx,
			accountID,
		)

	if err != nil {
		return nil, err
	}

	_ = cache.Set(
		ctx,
		s.redis,
		key,
		entriesDB,
		2*time.Minute,
	)

	return entriesDB, nil
}

func (s *Service) GetUserTotalCredit(
	ctx context.Context,
	userID uuid.UUID,
) (decimal.Decimal, error) {

	key := cache.UserTotalCreditKey(userID)

	value, err :=
		cache.Get[string](
			ctx,
			s.redis,
			key,
		)

	if err == nil && value != nil {
		return decimal.NewFromString(*value)
	}

	total, err :=
		s.repo.GetUserTotalCredit(
			ctx,
			userID,
		)

	if err != nil {
		return decimal.Zero, err
	}

	_ = cache.Set(
		ctx,
		s.redis,
		key,
		total.String(),
		5*time.Minute,
	)

	return total, nil
}

func (s *Service) GetUserTotalDebit(
	ctx context.Context,
	userID uuid.UUID,
) (decimal.Decimal, error) {

	key := cache.UserTotalDebitKey(userID)

	value, err :=
		cache.Get[string](
			ctx,
			s.redis,
			key,
		)

	if err == nil && value != nil {
		return decimal.NewFromString(*value)
	}

	total, err :=
		s.repo.GetUserTotalDebit(
			ctx,
			userID,
		)

	if err != nil {
		return decimal.Zero, err
	}

	_ = cache.Set(
		ctx,
		s.redis,
		key,
		total.String(),
		5*time.Minute,
	)

	return total, nil
}

func (s *Service) GetUserTotalBalance(
	ctx context.Context,
	userID uuid.UUID,
) (decimal.Decimal, error) {

	key := cache.UserTotalBalanceKey(userID)

	value, err :=
		cache.Get[string](
			ctx,
			s.redis,
			key,
		)

	if err == nil && value != nil {
		return decimal.NewFromString(*value)
	}

	total, err :=
		s.repo.GetUserTotalBalance(
			ctx,
			userID,
		)

	if err != nil {
		return decimal.Zero, err
	}

	_ = cache.Set(
		ctx,
		s.redis,
		key,
		total.String(),
		5*time.Minute,
	)

	return total, nil
}

func (r *LedgerRepository) CreateClearingAccount(
	ctx context.Context,
	userID uuid.UUID,
	cardName string,
) (uuid.UUID, error) {

	accountName := fmt.Sprintf("%s Clearing", cardName)

	return r.CreateAccount(
		ctx,
		userID,
		accountName,
		CLEARING,
	)
}
