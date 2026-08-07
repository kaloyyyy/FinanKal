package cache

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func Delete(
	ctx context.Context,
	rdb *redis.Client,
	keys ...string,
) error {

	if rdb == nil {
		log.Println("redis disabled: skip cache delete")
		return nil
	}

	if len(keys) == 0 {
		return nil
	}

	log.Printf(
		"redis delete keys: %v",
		keys,
	)

	err := rdb.Del(ctx, keys...).Err()

	if err != nil {
		log.Printf(
			"redis delete failed: %v keys=%v",
			err,
			keys,
		)
		return err
	}

	log.Printf(
		"redis delete success: %d keys",
		len(keys),
	)

	return nil
}

func InvalidateAccount(
	ctx context.Context,
	rdb *redis.Client,
	accountID uuid.UUID,
) error {

	log.Printf(
		"invalidate account cache: %s",
		accountID,
	)

	return Delete(
		ctx,
		rdb,

		AccountBalanceKey(accountID),
		AccountSummaryKey(accountID),
		AccountLedgerEntriesKey(accountID),
	)
}

func InvalidateUser(
	ctx context.Context,
	rdb *redis.Client,
	userID uuid.UUID,
) error {

	log.Printf(
		"invalidate user cache: %s",
		userID,
	)

	return Delete(
		ctx,
		rdb,

		UserBalanceKey(userID),
		UserTotalCreditKey(userID),
		UserTotalDebitKey(userID),
		UserNetWorthKey(userID),

		UserCreditCardsKey(userID),
	)
}

func Invalidate(
	ctx context.Context,
	rdb *redis.Client,
	keys ...string,
) error {

	return Delete(
		ctx,
		rdb,
		keys...,
	)
}

func InvalidateCreditCard(
	ctx context.Context,
	rdb *redis.Client,
	cardID uuid.UUID,
	userID uuid.UUID,
) error {

	log.Printf(
		"invalidate credit card cache card=%s user=%s",
		cardID,
		userID,
	)

	return Invalidate(
		ctx,
		rdb,

		CreditCardKey(cardID),
		UserCreditCardsKey(userID),

		AccountBalanceKey(cardID),
		UserBalanceKey(userID),
	)
}

func InvalidateTransaction(
	ctx context.Context,
	rdb *redis.Client,
	userID uuid.UUID,
	accountIDs ...uuid.UUID,
) error {

	log.Printf(
		"invalidate transaction cache user=%s accounts=%v",
		userID,
		accountIDs,
	)

	for _, accountID := range accountIDs {

		if err := InvalidateAccount(
			ctx,
			rdb,
			accountID,
		); err != nil {
			return err
		}
	}

	return InvalidateUser(
		ctx,
		rdb,
		userID,
	)
}

func InvalidateCreditCardPayment(
	ctx context.Context,
	rdb *redis.Client,
	cardID uuid.UUID,
	statementID uuid.UUID,
	userID uuid.UUID,
	accountIDs ...uuid.UUID,
) error {

	log.Printf(
		"invalidate credit card payment cache card=%s statement=%s user=%s accounts=%v",
		cardID,
		statementID,
		userID,
		accountIDs,
	)

	keys := []string{
		CreditCardKey(cardID),
		CreditCardStatementKey(statementID),
		UserCreditCardsKey(userID),

		UserBalanceKey(userID),
		UserTotalCreditKey(userID),
		UserTotalDebitKey(userID),
		UserNetWorthKey(userID),
	}

	for _, accountID := range accountIDs {
		keys = append(
			keys,
			AccountBalanceKey(accountID),
			AccountSummaryKey(accountID),
			AccountLedgerEntriesKey(accountID),
		)
	}

	return Delete(
		ctx,
		rdb,
		keys...,
	)
}
