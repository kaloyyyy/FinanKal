package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

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
	InvalidateLedgerCache(ctx, s.redis, accountIDs...)

	// Invalidate user total caches via method (handles nil redis)
	s.InvalidateUserTotalCache(ctx, userID)

	return txID, nil
}

func (s *Service) GetBalance(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	key := fmt.Sprintf("balance:%s", accountID)

	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			balance, err := decimal.NewFromString(val)
			if err != nil {
				return decimal.Zero, err
			}
			log.Printf("cache hit for balance: %s", accountID)
			return balance, nil
		} else if !errors.Is(err, redis.Nil) {
			// real Redis error
			return decimal.Zero, err
		}
	}

	balance, err := s.repo.GetBalance(ctx, accountID)
	if err != nil {
		return decimal.Zero, err
	}

	if s.redis != nil {
		if err := s.redis.Set(ctx, key, balance.String(), 5*time.Minute).Err(); err != nil {
			log.Printf("redis set error for key %s: %v", key, err)
		}
	}

	return balance, nil
}

func (s *Service) GetAccountSummary(ctx context.Context, accountID uuid.UUID) (*AccountSummary, error) {
	key := fmt.Sprintf("account_summary:%s", accountID)

	// 🔹 Try Redis first
	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var summary AccountSummary
			if err := json.Unmarshal([]byte(val), &summary); err == nil {
				log.Printf("cache hit for account summary: %s", accountID)
				return &summary, nil
			}
			// if unmarshal fails → fall through to DB
		} else if !errors.Is(err, redis.Nil) {
			// real Redis error
			return nil, err
		}
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
		if s.redis != nil {
			if err := s.redis.Set(ctx, key, data, 10*time.Minute).Err(); err != nil {
				log.Printf("redis set error for key %s: %v", key, err)
			}
		}
	}

	return summary, nil
}

func (s *Service) GetLedgerEntries(ctx context.Context, accountID uuid.UUID) ([]Entry, error) {
	key := fmt.Sprintf("ledger_entries:%s", accountID)

	// Try Redis first
	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var entries []Entry
			if err := json.Unmarshal([]byte(val), &entries); err == nil {
				log.Printf("cache hit for ledger entries: %s", accountID)
				return entries, nil
			}
		} else if !errors.Is(err, redis.Nil) {
			// real Redis error
			return nil, err
		}
	}

	// Fetch from DB
	entries, err := s.repo.GetLedgerEntries(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Cache result
	if data, err := json.Marshal(entries); err == nil {
		if s.redis != nil {
			if err := s.redis.Set(ctx, key, data, 2*time.Minute).Err(); err != nil {
				log.Printf("redis set error for key %s: %v", key, err)
			}
		}
	}

	return entries, nil
}

func (s *Service) GetUserTotalCredit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	key := fmt.Sprintf("user_total_credit:%s", userID)

	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			total, err := decimal.NewFromString(val)
			if err != nil {
				return decimal.Zero, err
			}
			log.Printf("cache hit for user total credit: %s", userID)
			return total, nil
		} else if !errors.Is(err, redis.Nil) {
			return decimal.Zero, err
		}
	}

	total, err := s.repo.GetUserTotalCredit(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}

	if s.redis != nil {
		if err := s.redis.Set(ctx, key, total.String(), 5*time.Minute).Err(); err != nil {
			log.Printf("redis set error for key %s: %v", key, err)
		}
	}

	return total, nil
}

func (s *Service) GetUserTotalDebit(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	key := fmt.Sprintf("user_total_debit:%s", userID)

	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			total, err := decimal.NewFromString(val)
			if err != nil {
				return decimal.Zero, err
			}
			log.Printf("cache hit for user total debit: %s", userID)
			return total, nil
		} else if !errors.Is(err, redis.Nil) {
			return decimal.Zero, err
		}
	}

	total, err := s.repo.GetUserTotalDebit(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}

	if s.redis != nil {
		if err := s.redis.Set(ctx, key, total.String(), 5*time.Minute).Err(); err != nil {
			log.Printf("redis set error for key %s: %v", key, err)
		}
	}

	return total, nil
}

func (s *Service) GetUserTotalBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	key := fmt.Sprintf("user_total_balance:%s", userID)

	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			total, err := decimal.NewFromString(val)
			if err != nil {
				return decimal.Zero, err
			}
			log.Printf("cache hit for user total balance: %s", userID)
			return total, nil
		} else if !errors.Is(err, redis.Nil) {
			return decimal.Zero, err
		}
	}

	total, err := s.repo.GetUserTotalBalance(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}

	if s.redis != nil {
		if err := s.redis.Set(ctx, key, total.String(), 5*time.Minute).Err(); err != nil {
			log.Printf("redis set error for key %s: %v", key, err)
		}
	}

	return total, nil
}
