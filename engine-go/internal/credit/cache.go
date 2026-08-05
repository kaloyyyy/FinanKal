package credit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/internal/cache"
)

const creditCardCacheTTL = 30 * time.Minute

func (s *Service) InvalidateCreditCardCache(
	ctx context.Context,
	cardID uuid.UUID,
) error {

	if s.redis == nil {
		return nil
	}

	return s.redis.Del(
		ctx,
		cache.CreditCardKey(cardID),
	).Err()
}

func (s *Service) InvalidateUserCreditCardsCache(
	ctx context.Context,
	userID uuid.UUID,
) error {

	if s.redis == nil {
		return nil
	}

	return s.redis.Del(
		ctx,
		cache.UserCreditCardsKey(userID),
	).Err()
}

func (s *Service) GetCreditCardCache(
	ctx context.Context,
	cardID uuid.UUID,
) (*CreditCard, error) {

	if s.redis == nil {
		return nil, nil
	}

	data, err := s.redis.Get(
		ctx,
		cache.CreditCardKey(cardID),
	).Bytes()

	if err != nil {
		return nil, err
	}

	var card CreditCard

	err = json.Unmarshal(
		data,
		&card,
	)

	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (s *Service) SetCreditCardCache(
	ctx context.Context,
	card *CreditCard,
) error {

	if s.redis == nil {
		return nil
	}

	data, err := json.Marshal(card)

	if err != nil {
		return err
	}

	return s.redis.Set(
		ctx,
		cache.CreditCardKey(card.ID),
		data,
		creditCardCacheTTL,
	).Err()
}

func (s *Service) GetUserCreditCardsCache(
	ctx context.Context,
	userID uuid.UUID,
) ([]CreditCard, error) {

	if s.redis == nil {
		return nil, nil
	}

	data, err := s.redis.Get(
		ctx,
		cache.UserCreditCardsKey(userID),
	).Bytes()

	if err != nil {
		return nil, err
	}

	var cards []CreditCard

	err = json.Unmarshal(
		data,
		&cards,
	)

	if err != nil {
		return nil, err
	}

	return cards, nil
}

func (s *Service) SetUserCreditCardsCache(
	ctx context.Context,
	userID uuid.UUID,
	cards []CreditCard,
) error {

	if s.redis == nil {
		return nil
	}

	data, err := json.Marshal(cards)

	if err != nil {
		return err
	}

	return s.redis.Set(
		ctx,
		cache.UserCreditCardsKey(userID),
		data,
		creditCardCacheTTL,
	).Err()
}

func (s *Service) InvalidateCreditCardTransactionCache(
	ctx context.Context,
	cardAccountID uuid.UUID,
	expenseAccountID uuid.UUID,
	userID uuid.UUID,
) {

	if s.redis == nil {
		return
	}

	keys := []string{
		fmt.Sprintf("balance:%s", cardAccountID),
		fmt.Sprintf("balance:%s", expenseAccountID),

		fmt.Sprintf("account_summary:%s", cardAccountID),
		fmt.Sprintf("account_summary:%s", expenseAccountID),

		fmt.Sprintf("ledger_entries:%s", cardAccountID),
		fmt.Sprintf("ledger_entries:%s", expenseAccountID),

		fmt.Sprintf("user_total_credit:%s", userID),
		fmt.Sprintf("user_total_debit:%s", userID),
		fmt.Sprintf("user_total_balance:%s", userID),
	}

	s.redis.Del(
		ctx,
		keys...,
	)
}
