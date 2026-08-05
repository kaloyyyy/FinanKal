package credit

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidPaymentAmount = errors.New(
		"payment amount must be greater than zero",
	)

	ErrStatementAlreadyPaid = errors.New(
		"statement already paid",
	)

	ErrPaymentTooLarge = errors.New(
		"payment exceeds statement balance",
	)
)

func (s *Service) PayCreditCardStatement(
	ctx context.Context,
	request CreditCardPaymentRequest,
) (uuid.UUID, error) {

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		return uuid.Nil, ErrInvalidPaymentAmount
	}

	statement, err :=
		s.statementRepo.GetStatement(
			ctx,
			request.StatementID,
		)

	if err != nil {
		return uuid.Nil, err
	}

	if statement.Status == StatementPaid {
		return uuid.Nil, ErrStatementAlreadyPaid
	}

	if request.Amount.GreaterThan(statement.TotalAmount) {
		return uuid.Nil, ErrPaymentTooLarge
	}

	card, err :=
		s.cardRepo.GetCreditCard(
			ctx,
			request.CardID,
		)

	if err != nil {
		return uuid.Nil, err
	}

	userID, err :=
		s.ledgerRepo.GetUserIDFromAccount(
			ctx,
			card.AccountID,
		)

	if err != nil {
		return uuid.Nil, err
	}

	tx, err :=
		s.ledgerRepo.BeginTx(ctx)

	if err != nil {
		return uuid.Nil, err
	}

	defer tx.Rollback(ctx)

	transactionID, err :=
		s.ledgerRepo.CreateTransactionTx(
			ctx,
			tx,
			userID,
			request.Description,
		)

	if err != nil {
		return uuid.Nil, err
	}

	// Debit Credit Card Liability
	//
	// reduces what you owe

	_, err =
		s.ledgerRepo.InsertEntryTx(
			ctx,
			tx,
			transactionID,
			card.AccountID,
			request.Amount,
			"DEBIT",
		)

	if err != nil {
		return uuid.Nil, err
	}

	// Credit Cash/Bank
	//
	// money leaves your bank

	_, err =
		s.ledgerRepo.InsertEntryTx(
			ctx,
			tx,
			transactionID,
			request.PaymentAccountID,
			request.Amount,
			"CREDIT",
		)

	if err != nil {
		return uuid.Nil, err
	}

	remaining :=
		statement.TotalAmount.Sub(
			request.Amount,
		)

	if remaining.Equal(decimal.Zero) {

		err =
			s.statementRepo.UpdateStatementStatusTx(
				ctx,
				tx,
				request.StatementID,
				StatementPaid,
			)

		if err != nil {
			return uuid.Nil, err
		}

	} else {

		err =
			s.statementRepo.UpdateStatementTotalTx(
				ctx,
				tx,
				request.StatementID,
				remaining,
			)

		if err != nil {
			return uuid.Nil, err
		}
	}

	err =
		tx.Commit(ctx)

	if err != nil {
		return uuid.Nil, err
	}

	s.InvalidateCreditCardTransactionCache(
		ctx,
		card.AccountID,
		request.PaymentAccountID,
		userID,
	)

	return transactionID, nil
}
