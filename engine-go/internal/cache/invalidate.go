package cache

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func accountKeys(accountID uuid.UUID) []string {
	return []string{
		AccountBalanceKey(accountID),
		AccountSummaryKey(accountID),
		AccountLedgerEntriesKey(accountID),
	}
}

func userKeys(userID uuid.UUID) []string {
	return []string{
		UserAccountsKey(userID),
		UserTotalCreditKey(userID),
		UserTotalDebitKey(userID),
		UserNetWorthKey(userID),
		UserCreditCardsKey(userID),
	}
}

// deleteKeys logs failures instead of returning them: the write being invalidated has
// already committed, so a stale cache must not fail the request. TTLs bound the staleness.
func deleteKeys(ctx context.Context, rdb *redis.Client, keys []string) {
	if rdb == nil {
		return
	}

	if err := rdb.Del(ctx, keys...).Err(); err != nil {
		log.Printf("cache delete failed: %v keys=%v", err, keys)
	}
}

// InvalidateTransaction clears everything derived from the balances of accountIDs.
func InvalidateTransaction(
	ctx context.Context,
	rdb *redis.Client,
	userID uuid.UUID,
	accountIDs ...uuid.UUID,
) {
	keys := userKeys(userID)
	for _, accountID := range accountIDs {
		keys = append(keys, accountKeys(accountID)...)
	}

	deleteKeys(ctx, rdb, keys)
}

// InvalidateCreditCard clears the card itself and everything derived from the balances
// of the card account and accountIDs.
func InvalidateCreditCard(
	ctx context.Context,
	rdb *redis.Client,
	cardID uuid.UUID,
	userID uuid.UUID,
	accountIDs ...uuid.UUID,
) {
	deleteKeys(ctx, rdb, []string{CreditCardKey(cardID)})
	InvalidateTransaction(ctx, rdb, userID, append(accountIDs, cardID)...)
}
