package credit

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

const DefaultUserID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

type Service struct {
	cardRepo      *CardRepository
	statementRepo *StatementRepository
	entryRepo     *EntryRepository
	ledgerRepo    *ledger.LedgerRepository
	redis         *redis.Client
	db            *pgxpool.Pool
}

func NewService(
	db *pgxpool.Pool,
	cardRepo *CardRepository,
	ledgerRepo *ledger.LedgerRepository,
	statementRepo *StatementRepository,
	entryRepo *EntryRepository,
	redis *redis.Client,
) *Service {

	return &Service{
		db:            db,
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
	accountName string,
	creditLimit decimal.Decimal,
	billingDay int,
	paymentDueDays int,
) (uuid.UUID, error) {

	// Validate inputs
	if billingDay < 1 || billingDay > 28 {
		return uuid.Nil, errors.New("billing day must be between 1 and 28")
	}

	if paymentDueDays < 1 || paymentDueDays > 28 {
		return uuid.Nil, errors.New("payment due days must be between 1 and 28")
	}

	if creditLimit.IsNegative() {
		return uuid.Nil, errors.New("credit limit cannot be negative")
	}

	var (
		userID            uuid.UUID
		clearingAccountID uuid.UUID
		err               error
	)

	// Create a new ledger account if none was supplied
	userID = uuid.MustParse(DefaultUserID)

	if accountID == uuid.Nil {

		accountID, err = s.ledgerRepo.CreateAccount(
			ctx,
			userID,
			accountName,
			ledger.CREDIT_CARD,
		)

		if err != nil {
			return uuid.Nil, err
		}

		clearingAccountID, err = s.ledgerRepo.CreateClearingAccount(
			ctx,
			userID,
			accountName,
		)
		if err != nil {
			return uuid.Nil, err
		}

	} else {

		userID, err = s.ledgerRepo.GetUserIDFromAccount(ctx, accountID)
		if err != nil {
			return uuid.Nil, err
		}

		accountType, err := s.cardRepo.GetAccountType(ctx, accountID)
		if err != nil {
			return uuid.Nil, err
		}

		if accountType != ledger.CREDIT_CARD {
			return uuid.Nil, errors.New("account is not a credit card account")
		}

		// Existing credit card account. Create its clearing account.
		clearingAccountID, err = s.ledgerRepo.CreateClearingAccount(
			ctx,
			userID,
			accountName,
		)
		if err != nil {
			return uuid.Nil, err
		}
	}

	cardID, err := s.cardRepo.CreateCreditCard(
		ctx,
		accountID,
		clearingAccountID,
		creditLimit,
		billingDay,
		paymentDueDays,
	)
	if err != nil {
		return uuid.Nil, err
	}

	_ = cache.InvalidateCreditCard(
		ctx,
		s.redis,
		cardID,
		userID,
	)

	return cardID, nil
}

func (s *Service) GetCreditCard(
	ctx context.Context,
	cardID uuid.UUID,
) (*CreditCard, error) {

	key := cache.CreditCardKey(cardID)

	// Try Redis cache first
	card, err := cache.Get[CreditCard](
		ctx,
		s.redis,
		key,
	)

	if err == nil && card != nil {
		return card, nil
	}

	// Cache miss -> database
	cardDB, err := s.cardRepo.GetCreditCard(
		ctx,
		cardID,
	)

	if err != nil {
		return nil, err
	}

	// Store in Redis
	// Cache failure should not fail request
	_ = cache.Set(
		ctx,
		s.redis,
		key,
		cardDB,
		30*time.Minute,
	)

	return cardDB, nil
}

func (s *Service) ListCreditCards(
	ctx context.Context,
	userID uuid.UUID,
) ([]CreditCard, error) {

	key := cache.UserCreditCardsKey(userID)

	cards, err := cache.Get[[]CreditCard](
		ctx,
		s.redis,
		key,
	)

	if err == nil && cards != nil {
		return *cards, nil
	}

	cardsDB, err := s.cardRepo.ListCreditCards(
		ctx,
		userID,
	)

	if err != nil {
		return nil, err
	}

	_ = cache.Set(
		ctx,
		s.redis,
		key,
		cardsDB,
		30*time.Minute,
	)

	return cardsDB, nil
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

	userID, err := s.ledgerRepo.GetUserIDFromAccount(ctx, card.CardID)
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

	_ = cache.InvalidateCreditCard(
		ctx,
		s.redis,
		cardID,
		userID,
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

	userID, err := s.ledgerRepo.GetUserIDFromAccount(ctx, card.CardID)
	if err != nil {
		return err
	}

	err = s.cardRepo.DeleteCreditCard(ctx, cardID)

	if err != nil {
		return err
	}

	_ = cache.InvalidateUser(
		ctx,
		s.redis,
		userID,
	)

	_ = cache.InvalidateCreditCard(
		ctx,
		s.redis,
		cardID,
		userID,
	)
	return nil
}

func (r *StatementRepository) UpdateStatementStatusTx(
	ctx context.Context,
	tx pgx.Tx,
	statementID uuid.UUID,
	status StatementStatus,
) error {

	_, err := tx.Exec(
		ctx,
		`
		UPDATE credit_card_statements
		SET status=$2,
		    updated_at=NOW()
		WHERE id=$1
		`,
		statementID,
		status,
	)

	return err
}
