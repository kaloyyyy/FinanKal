package credit

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/shopspring/decimal"
)

func (s *Service) RecordCreditCardTransaction(
	ctx context.Context,
	request CreditCardTransactionRequest,
) (uuid.UUID, error) {

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		return uuid.Nil, errors.New(
			"amount must be greater than zero",
		)
	}

	// Get credit card

	card, err := s.cardRepo.GetCreditCard(
		ctx,
		request.CardID,
	)

	if err != nil {
		return uuid.Nil, err
	}

	// Validate transaction

	err = s.ValidateCreditCardTransaction(
		ctx,
		request,
		*card,
	)

	if err != nil {
		return uuid.Nil, err
	}

	// Begin database transaction

	tx, err := s.ledgerRepo.BeginTx(ctx)

	if err != nil {
		return uuid.Nil, err
	}

	defer tx.Rollback(ctx)

	// Calculate billing cycle

	cycle := CalculateBillingCycle(
		request.PurchaseDate,
		*card,
	)

	// Find existing statement

	statement, err :=
		s.statementRepo.FindStatementByCycleTx(
			ctx,
			tx,
			request.CardID,
			cycle,
		)
	// Create statement if missing

	if err != nil {

		statementID, err :=
			s.statementRepo.CreateStatementTx(
				ctx,
				tx,
				CreditCardStatement{
					CreditCardID:  request.CardID,
					StartDate:     cycle.CycleStartDate,
					EndDate:       cycle.CycleEndDate,
					StatementDate: cycle.StatementDate,
					DueDate:       cycle.DueDate,
					TotalAmount:   decimal.Zero,
					Status:        StatementOpen,
				},
			)

		if err != nil {
			return uuid.Nil, err
		}

		statement, err =
			s.statementRepo.GetStatementTx(
				ctx,
				tx,
				statementID,
			)

		if err != nil {
			return uuid.Nil, err
		}
	}

	// Get owner

	userID, err :=
		s.ledgerRepo.GetUserIDFromAccount(
			ctx,
			card.AccountID,
		)

	if err != nil {
		return uuid.Nil, err
	}

	// Create ledger transaction

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

	// Debit expense

	_, err =
		s.ledgerRepo.InsertEntryTx(
			ctx,
			tx,
			transactionID,
			request.ExpenseAccountID,
			request.Amount,
			ledger.DEBIT,
		)

	if err != nil {
		return uuid.Nil, err
	}

	// Credit credit card liability

	cardEntryID, err :=
		s.ledgerRepo.InsertEntryTx(
			ctx,
			tx,
			transactionID,
			card.AccountID,
			request.Amount,
			ledger.CREDIT,
		)

	if err != nil {
		return uuid.Nil, err
	}

	// Attach entry to statement

	_, err =
		s.entryRepo.CreateEntryMappingTx(
			ctx,
			tx,
			statement.ID,
			cardEntryID,
		)

	if err != nil {
		return uuid.Nil, err
	}

	// Update statement total

	err =
		s.statementRepo.IncrementStatementTotalTx(
			ctx,
			tx,
			statement.ID,
			request.Amount,
		)

	if err != nil {
		return uuid.Nil, err
	}

	// Commit transaction

	err = tx.Commit(ctx)

	if err != nil {
		return uuid.Nil, err
	}

	// Redis invalidation

	if s.redis != nil {
		s.InvalidateCreditCardTransactionCache(
			ctx,
			card.AccountID,
			request.ExpenseAccountID,
			userID,
		)
	}

	return transactionID, nil
}
