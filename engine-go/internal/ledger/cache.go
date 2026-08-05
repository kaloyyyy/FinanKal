package ledger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func InvalidateLedgerCache(
	ctx context.Context,
	redis *redis.Client,
	accountIDs ...uuid.UUID,
) {

	if redis == nil {
		return
	}

	var keys []string

	for _, id := range accountIDs {

		keys = append(
			keys,
			fmt.Sprintf("balance:%s", id),
			fmt.Sprintf("account_summary:%s", id),
			fmt.Sprintf("ledger_entries:%s", id),
		)
	}

	redis.Del(ctx, keys...)
}

// =========================
// Cache Invalidation
// =========================

func (s *Service) InvalidateAccountCache(
	ctx context.Context,
	accountID uuid.UUID,
) {

	keys := []string{
		fmt.Sprintf("balance:%s", accountID),
		fmt.Sprintf("account_summary:%s", accountID),
		fmt.Sprintf("ledger_entries:%s", accountID),
	}

	s.redis.Del(
		ctx,
		keys...,
	)
}

func (s *Service) InvalidateUserTotalCache(
	ctx context.Context,
	userID uuid.UUID,
) {

	keys := []string{
		fmt.Sprintf("user_total_credit:%s", userID),
		fmt.Sprintf("user_total_debit:%s", userID),
		fmt.Sprintf("user_total_balance:%s", userID),
	}

	s.redis.Del(
		ctx,
		keys...,
	)
}
