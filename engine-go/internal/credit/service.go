package credit

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type Service struct {
	cardRepo      *CardRepository
	statementRepo *StatementRepository
	entryRepo     *EntryRepository
	ledgerRepo    *ledger.LedgerRepository
	redis         *redis.Client
}

func NewService(
	cardRepo *CardRepository, ledgerRepo *ledger.LedgerRepository, statementRepo *StatementRepository, entryRepo *EntryRepository, redis *redis.Client,
) *Service {
	return &Service{
		cardRepo:      cardRepo,
		ledgerRepo:    ledgerRepo,
		statementRepo: statementRepo,
		entryRepo:     entryRepo,
		redis:         redis,
	}
}

// =========================
// Credit Card CRUD
// =========================

func (s *Service) CreateCreditCard(
	ctx context.Context,
	accountID uuid.UUID,
	creditLimit decimal.Decimal,
	billingDay int,
	paymentDueDays int,
) (uuid.UUID, error) {
	userID, err := s.ledgerRepo.GetUserIDFromAccount(ctx, accountID)
	if err != nil {
		return uuid.Nil, err
	}
	accountType, err := s.cardRepo.GetAccountType(ctx, accountID)
	if err != nil {
		return uuid.Nil, err
	}

	if accountType != "CREDIT_CARD" {
		return uuid.Nil, errors.New("account is not a credit card account")
	}

	if billingDay < 1 || billingDay > 28 {
		return uuid.Nil, errors.New("billing day must be between 1 and 28")
	}

	if paymentDueDays < 1 || paymentDueDays > 28 {
		return uuid.Nil, errors.New("payment due days must be between 1 and 28")
	}

	if creditLimit.IsNegative() {
		return uuid.Nil, errors.New("credit limit cannot be negative")
	}

	cardID, err := s.cardRepo.CreateCreditCard(
		ctx,
		accountID,
		creditLimit,
		billingDay,
		paymentDueDays,
	)

	if err != nil {
		return uuid.Nil, err
	}

	s.InvalidateUserCreditCardsCache(
		ctx,
		userID,
	)

	s.InvalidateCreditCardCache(
		ctx,
		cardID,
	)

	return cardID, nil
}

func (s *Service) GetCreditCard(
	ctx context.Context,
	cardID uuid.UUID,
) (*CreditCard, error) {

	return s.cardRepo.GetCreditCard(ctx, cardID)
}

func (s *Service) ListCreditCards(
	ctx context.Context,
	userID uuid.UUID,
) ([]CreditCard, error) {

	return s.cardRepo.ListCreditCards(ctx, userID)
}

func (s *Service) UpdateCreditCard(
	ctx context.Context,
	cardID uuid.UUID,
	creditLimit decimal.Decimal,
	billingDay int,
	paymentDueDays int,
) error {
	card, err := s.cardRepo.GetCreditCard(ctx, cardID)
	if err != nil {
		return err
	}

	userID, err := s.ledgerRepo.GetUserIDFromAccount(ctx, card.AccountID)
	if err != nil {
		return err
	}
	if billingDay < 1 || billingDay > 28 {
		return errors.New("billing day must be between 1 and 28")
	}

	if paymentDueDays < 1 || paymentDueDays > 28 {
		return errors.New("payment due days must be between 1 and 28")
	}

	if creditLimit.IsNegative() {
		return errors.New("credit limit cannot be negative")
	}

	err = s.cardRepo.UpdateCreditCard(
		ctx,
		cardID,
		creditLimit,
		billingDay,
		paymentDueDays,
	)

	if err != nil {
		return err
	}

	s.InvalidateUserCreditCardsCache(
		ctx,
		userID,
	)

	s.InvalidateCreditCardCache(
		ctx,
		cardID,
	)

	return nil
}

func (s *Service) DeleteCreditCard(
	ctx context.Context,
	cardID uuid.UUID,
) error {

	card, err := s.cardRepo.GetCreditCard(ctx, cardID)
	if err != nil {
		return err
	}

	userID, err := s.ledgerRepo.GetUserIDFromAccount(ctx, card.AccountID)
	if err != nil {
		return err
	}

	err = s.cardRepo.DeleteCreditCard(ctx, cardID)

	if err != nil {
		return err
	}

	s.InvalidateUserCreditCardsCache(
		ctx,
		userID,
	)

	s.InvalidateCreditCardCache(
		ctx,
		cardID,
	)
	return nil
}
