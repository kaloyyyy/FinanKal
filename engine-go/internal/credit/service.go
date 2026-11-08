package credit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

// shortcut: new card accounts all belong to this user because CreateCreditCardRequest
// has no user_id. Add one to the proto once the API has real users.
const DefaultUserID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

type Service struct {
	cardRepo      *CardRepository
	statementRepo *StatementRepository
	entryRepo     *EntryRepository
	ledgerRepo    *ledger.LedgerRepository
	redis         *redis.Client
}

func NewService(
	cardRepo *CardRepository,
	ledgerRepo *ledger.LedgerRepository,
	statementRepo *StatementRepository,
	entryRepo *EntryRepository,
	redis *redis.Client,
) *Service {

	return &Service{
		cardRepo:      cardRepo,
		ledgerRepo:    ledgerRepo,
		statementRepo: statementRepo,
		entryRepo:     entryRepo,
		redis:         redis,
	}
}

// CreateCreditCard creates a card on accountID, or on a new account when accountID is uuid.Nil.
// Everything is created in one transaction, so a failure leaves no orphaned accounts.
func (s *Service) CreateCreditCard(
	ctx context.Context,
	accountID uuid.UUID,
	accountName string,
	creditLimit decimal.Decimal,
	billingDay int,
	paymentDueDays int,
) (uuid.UUID, error) {

	if err := validateCardTerms(creditLimit, billingDay, paymentDueDays); err != nil {
		return uuid.Nil, err
	}

	userID := uuid.MustParse(DefaultUserID)
	createAccount := accountID == uuid.Nil

	if createAccount {
		if strings.TrimSpace(accountName) == "" {
			return uuid.Nil, fmt.Errorf("%w: account name is required for a new card account", ErrInvalidRequest)
		}
	} else {
		var err error

		userID, err = s.ledgerRepo.GetUserIDFromAccount(ctx, accountID)
		if err != nil {
			return uuid.Nil, err
		}

		existingName, accountType, _, err := s.ledgerRepo.GetAccount(ctx, accountID)
		if err != nil {
			return uuid.Nil, err
		}
		if accountType != ledger.CREDIT_CARD {
			return uuid.Nil, ErrInvalidCreditAccount
		}
		if accountName == "" {
			accountName = existingName
		}
	}

	tx, err := s.ledgerRepo.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	if createAccount {
		accountID, err = s.ledgerRepo.CreateAccountTx(ctx, tx, userID, accountName, ledger.CREDIT_CARD)
		if err != nil {
			return uuid.Nil, err
		}
	}

	clearingAccountID, err := s.ledgerRepo.CreateClearingAccountTx(ctx, tx, userID, accountName)
	if err != nil {
		return uuid.Nil, err
	}

	cardID, err := s.cardRepo.CreateCreditCardTx(ctx, tx, accountID, clearingAccountID, creditLimit, billingDay, paymentDueDays)
	if err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}

	cache.InvalidateCreditCard(ctx, s.redis, cardID, userID, clearingAccountID)

	return cardID, nil
}

func (s *Service) GetCreditCard(ctx context.Context, cardID uuid.UUID) (*CreditCard, error) {
	return cache.GetOrLoad(ctx, s.redis, cache.CreditCardKey(cardID), 30*time.Minute, func() (*CreditCard, error) {
		return s.cardRepo.GetCreditCard(ctx, cardID)
	})
}

func (s *Service) ListCreditCards(ctx context.Context, userID uuid.UUID) ([]CreditCard, error) {
	return cache.GetOrLoad(ctx, s.redis, cache.UserCreditCardsKey(userID), 30*time.Minute, func() ([]CreditCard, error) {
		return s.cardRepo.ListCreditCards(ctx, userID)
	})
}

func (s *Service) UpdateCreditCard(
	ctx context.Context,
	cardID uuid.UUID,
	creditLimit decimal.Decimal,
	billingDay int,
	paymentDueDays int,
) error {

	if err := validateCardTerms(creditLimit, billingDay, paymentDueDays); err != nil {
		return err
	}

	userID, err := s.ledgerRepo.GetUserIDFromAccount(ctx, cardID)
	if err != nil {
		return err
	}

	if err := s.cardRepo.UpdateCreditCard(ctx, cardID, creditLimit, billingDay, paymentDueDays); err != nil {
		return err
	}

	cache.InvalidateCreditCard(ctx, s.redis, cardID, userID)

	return nil
}

func (s *Service) DeleteCreditCard(ctx context.Context, cardID uuid.UUID) error {
	userID, err := s.ledgerRepo.GetUserIDFromAccount(ctx, cardID)
	if err != nil {
		return err
	}

	if err := s.cardRepo.DeleteCreditCard(ctx, cardID); err != nil {
		return err
	}

	cache.InvalidateCreditCard(ctx, s.redis, cardID, userID)

	return nil
}
