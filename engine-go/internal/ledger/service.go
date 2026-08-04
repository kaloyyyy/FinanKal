package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/pgxpool"
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
		err := s.repo.InsertEntry(ctx, txID, e.AccountID, e.Amount, string(e.Type))
		if err != nil {
			return uuid.UUID{}, err
		}
	}

	for _, e := range entries {
		balanceKey := fmt.Sprintf("balance:%s", e.AccountID)
		summaryKey := fmt.Sprintf("account_summary:%s", e.AccountID)
		ledgerKey := fmt.Sprintf("ledger_entries:%s", e.AccountID)

		s.redis.Del(ctx, balanceKey)
		s.redis.Del(ctx, summaryKey)
		s.redis.Del(ctx, ledgerKey)
	}

	// Invalidate user total caches
	userCreditKey := fmt.Sprintf("user_total_credit:%s", userID)
	userDebitKey := fmt.Sprintf("user_total_debit:%s", userID)
	userBalanceKey := fmt.Sprintf("user_total_balance:%s", userID)
	s.redis.Del(ctx, userCreditKey)
	s.redis.Del(ctx, userDebitKey)
	s.redis.Del(ctx, userBalanceKey)

	return txID, nil
}

func (s *Service) GetBalance(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	key := fmt.Sprintf("balance:%s", accountID)

	val, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		balance, err := decimal.NewFromString(val)
		if err != nil {
			return decimal.Zero, err
		}
		return balance, nil
	}

	balance, err := s.repo.GetBalance(ctx, accountID)
	if err != nil {
		return decimal.Zero, err
	}

	s.redis.Set(ctx, key, balance.String(), 5*time.Minute)

	return balance, nil
}

func (s *Service) GetAccountSummary(ctx context.Context, accountID uuid.UUID) (*AccountSummary, error) {
	key := fmt.Sprintf("account_summary:%s", accountID)

	// 🔹 Try Redis first
	val, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		var summary AccountSummary
		if err := json.Unmarshal([]byte(val), &summary); err == nil {
			fmt.Println("Cache hit for account summary:", accountID)
			return &summary, nil
		}
		// if unmarshal fails → fall through to DB
	} else if !errors.Is(err, redis.Nil) {
		// real Redis error
		return nil, err
	}

	// 🔹 Fetch from DB
	name, accountType, createdAt, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	balance, err := s.repo.GetBalance(ctx, accountID)
	if err != nil {
		return nil, err
	}

	summary := &AccountSummary{
		AccountID: accountID,
		Name:      name,
		Type:      accountType,
		Balance:   balance,
		CreatedAt: createdAt,
	}

	// 🔹 Cache result (ignore cache error for now)
	if data, err := json.Marshal(summary); err == nil {
		s.redis.Set(ctx, key, data, 10*time.Minute)
	}

	return summary, nil
}

func (s *Service) GetLedgerEntries(ctx context.Context, accountID uuid.UUID) ([]Entry, error) {
	key := fmt.Sprintf("ledger_entries:%s", accountID)

	// Try Redis first
	val, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		var entries []Entry
		if err := json.Unmarshal([]byte(val), &entries); err == nil {
			fmt.Println("Cache hit for ledger entries:", accountID)
			return entries, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		// real Redis error
		return nil, err
	}

	// Fetch from DB
	entries, err := s.repo.GetLedgerEntries(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Cache result
	if data, err := json.Marshal(entries); err == nil {
		s.redis.Set(ctx, key, data, 2*time.Minute)
	}

	return entries, nil
}

func (s *Service) GetUserTotalCredit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	key := fmt.Sprintf("user_total_credit:%s", userID)

	val, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		total, err := decimal.NewFromString(val)
		if err != nil {
			return decimal.Zero, err
		}
		fmt.Println("Cache hit for user total credit:", userID)
		return total, nil
	}

	total, err := s.repo.GetUserTotalCredit(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}

	s.redis.Set(ctx, key, total.String(), 5*time.Minute)

	return total, nil
}

func (s *Service) GetUserTotalDebit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	key := fmt.Sprintf("user_total_debit:%s", userID)

	val, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		total, err := decimal.NewFromString(val)
		if err != nil {
			return decimal.Zero, err
		}
		fmt.Println("Cache hit for user total debit:", userID)
		return total, nil
	}

	total, err := s.repo.GetUserTotalDebit(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}

	s.redis.Set(ctx, key, total.String(), 5*time.Minute)

	return total, nil
}

func (s *Service) GetUserTotalBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	key := fmt.Sprintf("user_total_balance:%s", userID)

	val, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		total, err := decimal.NewFromString(val)
		if err != nil {
			return decimal.Zero, err
		}
		fmt.Println("Cache hit for user total balance:", userID)
		return total, nil
	}

	total, err := s.repo.GetUserTotalBalance(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}

	s.redis.Set(ctx, key, total.String(), 5*time.Minute)

	return total, nil
}
