package credit

import (
	"context"
	"errors"

	_ "github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/kaloy/finankal/engine-go/internal/ledger"
)

var (
	ErrInvalidAmount         = errors.New("amount must be greater than zero")
	ErrCreditCardNotFound    = errors.New("credit card not found")
	ErrInvalidCreditAccount  = errors.New("account is not a credit card")
	ErrCreditLimitExceeded   = errors.New("credit limit exceeded")
	ErrInvalidExpenseAccount = errors.New("invalid expense account")
)

func (s *Service) ValidateCreditCardTransaction(
	ctx context.Context,
	request CreditCardTransactionRequest,
	card CreditCard,
) error {

	// Validate amount
	if request.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	// Validate credit card account
	accountName, accountType, _, err := s.ledgerRepo.GetAccount(
		ctx,
		card.AccountID,
	)

	if err != nil {
		return ErrCreditCardNotFound
	}

	if accountType != ledger.CREDIT_CARD {
		return ErrInvalidCreditAccount
	}

	_ = accountName

	// Validate expense account
	_, expenseType, _, err := s.ledgerRepo.GetAccount(
		ctx,
		request.ExpenseAccountID,
	)

	if err != nil {
		return ErrInvalidExpenseAccount
	}

	if expenseType != ledger.EXPENSE {
		return ErrInvalidExpenseAccount
	}

	// Validate available credit limit
	currentBalance, err := s.ledgerRepo.GetBalance(
		ctx,
		card.AccountID,
	)

	if err != nil {
		return err
	}

	newBalance := currentBalance.Add(request.Amount)

	if newBalance.GreaterThan(card.CreditLimit) {
		return ErrCreditLimitExceeded
	}

	return nil
}
