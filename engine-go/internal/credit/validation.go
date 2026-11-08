package credit

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/shopspring/decimal"
)

var (
	// ErrInvalidRequest is wrapped by every error caused by bad input.
	ErrInvalidRequest = errors.New("invalid request")

	ErrInvalidAmount          = fmt.Errorf("%w: amount must be greater than zero with at most 2 decimal places", ErrInvalidRequest)
	ErrInvalidCreditAccount   = fmt.Errorf("%w: account is not a credit card", ErrInvalidRequest)
	ErrInvalidClearingAccount = fmt.Errorf("%w: invalid clearing account", ErrInvalidRequest)

	ErrCreditLimitExceeded = errors.New("credit limit exceeded")
)

// validateAmount rejects sub-cent amounts because credit_card_statements.total_amount
// is numeric(18,2): Postgres would round them, and the statement would drift from the ledger.
func validateAmount(amount decimal.Decimal) error {
	if !amount.IsPositive() || !amount.Equal(amount.Round(2)) {
		return ErrInvalidAmount
	}
	return nil
}

func validateCardTerms(creditLimit decimal.Decimal, billingDay, paymentDueDays int) error {
	if billingDay < 1 || billingDay > 28 {
		return fmt.Errorf("%w: billing day must be between 1 and 28", ErrInvalidRequest)
	}
	if paymentDueDays < 1 || paymentDueDays > 28 {
		return fmt.Errorf("%w: payment due days must be between 1 and 28", ErrInvalidRequest)
	}
	if creditLimit.IsNegative() {
		return fmt.Errorf("%w: credit limit cannot be negative", ErrInvalidRequest)
	}
	return nil
}

// ValidateCreditCardTransaction checks the amount and that the card's accounts have the right types.
func (s *Service) ValidateCreditCardTransaction(
	ctx context.Context,
	request CreditCardTransactionRequest,
	card CreditCard,
) error {
	if err := validateAmount(request.Amount); err != nil {
		return err
	}

	_, accountType, _, err := s.ledgerRepo.GetAccount(ctx, card.CardID)
	if err != nil {
		return err
	}
	if accountType != ledger.CREDIT_CARD {
		return ErrInvalidCreditAccount
	}

	_, clearingType, _, err := s.ledgerRepo.GetAccount(ctx, card.ClearingAccountId)
	if err != nil {
		return err
	}
	if clearingType != ledger.CLEARING {
		return ErrInvalidClearingAccount
	}

	return nil
}

// checkCreditLimit locks the card row until tx ends, so concurrent purchases on one
// card cannot each pass the check against the same balance.
func (s *Service) checkCreditLimit(
	ctx context.Context,
	tx pgx.Tx,
	card CreditCard,
	amount decimal.Decimal,
) error {
	limit, err := s.cardRepo.LockCreditLimitTx(ctx, tx, card.CardID)
	if err != nil {
		return err
	}

	balance, err := s.ledgerRepo.GetBalanceTx(ctx, tx, card.CardID)
	if err != nil {
		return err
	}

	// GetBalance is debits minus credits, so money owed on a card is a negative balance.
	if balance.Neg().Add(amount).GreaterThan(limit) {
		return ErrCreditLimitExceeded
	}

	return nil
}
