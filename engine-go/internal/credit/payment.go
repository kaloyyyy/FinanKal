package credit

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
)

var (
	ErrStatementAlreadyPaid = errors.New("statement already paid")
	ErrPaymentTooLarge      = errors.New("payment exceeds statement balance")

	ErrStatementCardMismatch = fmt.Errorf("%w: statement does not belong to this card", ErrInvalidRequest)
	ErrInvalidPaymentAccount = fmt.Errorf("%w: payment account must be an asset account owned by the cardholder", ErrInvalidRequest)
)

func (s *Service) PayCreditCardStatement(
	ctx context.Context,
	request CreditCardPaymentRequest,
) (uuid.UUID, error) {

	if err := validateAmount(request.Amount); err != nil {
		return uuid.Nil, err
	}

	card, err := s.cardRepo.GetCreditCard(ctx, request.CardID)
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := s.ledgerRepo.GetUserIDFromAccount(ctx, card.CardID)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.validatePaymentAccount(ctx, request.PaymentAccountID, userID); err != nil {
		return uuid.Nil, err
	}

	tx, err := s.ledgerRepo.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	// Locked until commit, so two concurrent payments cannot both pass the balance check.
	statement, err := s.statementRepo.GetStatementForUpdateTx(ctx, tx, request.StatementID)
	if err != nil {
		return uuid.Nil, err
	}

	if statement.CardID != request.CardID {
		return uuid.Nil, ErrStatementCardMismatch
	}
	if statement.Status == StatementPaid {
		return uuid.Nil, ErrStatementAlreadyPaid
	}
	if request.Amount.GreaterThan(statement.TotalAmount) {
		return uuid.Nil, ErrPaymentTooLarge
	}

	transactionID, err := s.ledgerRepo.CreateTransactionTx(ctx, tx, userID, request.Description, nil)
	if err != nil {
		return uuid.Nil, err
	}

	// Debiting the card reduces what is owed; crediting the bank is the money leaving.
	if _, err := s.ledgerRepo.InsertEntryTx(ctx, tx, transactionID, card.CardID, request.Amount, ledger.DEBIT); err != nil {
		return uuid.Nil, err
	}
	if _, err := s.ledgerRepo.InsertEntryTx(ctx, tx, transactionID, request.PaymentAccountID, request.Amount, ledger.CREDIT); err != nil {
		return uuid.Nil, err
	}

	remaining := statement.TotalAmount.Sub(request.Amount)
	if remaining.IsZero() {
		err = s.statementRepo.UpdateStatementStatusTx(ctx, tx, request.StatementID, StatementPaid)
	} else {
		err = s.statementRepo.UpdateStatementTotalTx(ctx, tx, request.StatementID, remaining)
	}
	if err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}

	cache.InvalidateCreditCard(ctx, s.redis, card.CardID, userID, request.PaymentAccountID)

	return transactionID, nil
}

func (s *Service) validatePaymentAccount(ctx context.Context, accountID, userID uuid.UUID) error {
	owner, err := s.ledgerRepo.GetUserIDFromAccount(ctx, accountID)
	if err != nil {
		return err
	}

	_, accountType, _, err := s.ledgerRepo.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}

	if owner != userID || accountType != ledger.ASSET {
		return ErrInvalidPaymentAccount
	}

	return nil
}
