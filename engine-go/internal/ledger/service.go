package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kaloy/finankal/engine-go/internal/model"
	"github.com/kaloy/finankal/engine-go/internal/repository"
	"github.com/shopspring/decimal"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo  *repository.LedgerRepository
	redis *redis.Client
}

func NewService(repo *repository.LedgerRepository, redis *redis.Client) *Service {
	return &Service{repo: repo, redis: redis}
}

func (s *Service) CreateTransaction(ctx context.Context, description string, entries []model.Entry) (uuid.UUID, error) {
	if err := validateEntries(entries); err != nil {
		return uuid.UUID{}, err
	}

	txID, err := s.repo.CreateTransaction(ctx, description)
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

		s.redis.Del(ctx, balanceKey)
		s.redis.Del(ctx, summaryKey)
	}

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

	s.redis.Set(ctx, key, balance.String(), 0)

	return balance, nil
}

func (s *Service) GetAccountSummary(ctx context.Context, accountID uuid.UUID) (*model.AccountSummary, error) {
	key := fmt.Sprintf("account_summary:%s", accountID)

	// 🔹 Try Redis first
	val, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		var summary model.AccountSummary
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

	summary := &model.AccountSummary{
		AccountID: accountID,
		Name:      name,
		Type:      accountType,
		Balance:   balance,
		CreatedAt: createdAt,
	}

	// 🔹 Cache result (ignore cache error for now)
	if data, err := json.Marshal(summary); err == nil {
		s.redis.Set(ctx, key, data, 0)
	}

	return summary, nil
}
