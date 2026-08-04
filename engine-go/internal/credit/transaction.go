package credit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/shopspring/decimal"
)

type CreditCardTransactionRequest struct {
	UserID           uuid.UUID
	CreditCardID     uuid.UUID
	ExpenseAccountID uuid.UUID
	Amount           decimal.Decimal
	Description      string
	PurchaseDate     time.Time
}

func (s *Service) RecordCreditCardTransaction(
	ctx context.Context,
	request CreditCardTransactionRequest,
) (uuid.UUID, error) {

	tx, err := s.ledgerRepo.BeginTx(ctx)

	if err != nil {
		return uuid.Nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	card, err := s.cardRepo.GetCreditCard(
		ctx,
		request.CreditCardID,
	)

	if err != nil {
		return uuid.Nil, err
	}

	err = s.ValidateCreditCardTransaction(
		ctx,
		request,
		*card,
	)

	if err != nil {
		return uuid.Nil, err
	}

	statement, err := s.GetOrCreateStatement(
		ctx,
		*card,
		request.PurchaseDate,
	)

	if err != nil {
		return uuid.Nil, err
	}

	transactionID, err :=
		s.ledgerRepo.CreateTransactionTx(
			ctx,
			tx,
			request.UserID,
			request.Description,
		)

	if err != nil {
		return uuid.Nil, err
	}

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

	creditEntryID, err :=
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

	_, err =
		s.entryRepo.CreateEntryMappingTx(
			ctx,
			tx,
			statement.ID,
			creditEntryID,
		)

	if err != nil {
		return uuid.Nil, err
	}

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

	err = tx.Commit(ctx)

	if err != nil {
		return uuid.Nil, err
	}

	s.InvalidateUserCreditCardsCache(
		ctx,
		request.UserID,
	)

	return transactionID, nil
}
