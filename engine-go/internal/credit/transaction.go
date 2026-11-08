package credit

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/shopspring/decimal"
)

func (s *Service) RecordCreditCardTransaction(
	ctx context.Context,
	request CreditCardTransactionRequest,
) (uuid.UUID, error) {

	card, err := s.cardRepo.GetCreditCard(ctx, request.CardID)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.ValidateCreditCardTransaction(ctx, request, *card); err != nil {
		return uuid.Nil, err
	}

	userID, err := s.ledgerRepo.GetUserIDFromAccount(ctx, card.CardID)
	if err != nil {
		return uuid.Nil, err
	}

	tx, err := s.ledgerRepo.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	if err := s.checkCreditLimit(ctx, tx, *card, request.Amount); err != nil {
		return uuid.Nil, err
	}

	cycle := CalculateBillingCycle(request.PurchaseDate, *card)

	statementID, err := s.findOrCreateStatementTx(ctx, tx, card.CardID, cycle)
	if err != nil {
		return uuid.Nil, err
	}

	transactionID, err := s.ledgerRepo.CreateTransactionTx(ctx, tx, userID, request.Description)
	if err != nil {
		return uuid.Nil, err
	}

	if _, err := s.ledgerRepo.InsertEntryTx(ctx, tx, transactionID, card.ClearingAccountId, request.Amount, ledger.DEBIT); err != nil {
		return uuid.Nil, err
	}

	cardEntryID, err := s.ledgerRepo.InsertEntryTx(ctx, tx, transactionID, card.CardID, request.Amount, ledger.CREDIT)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.entryRepo.CreateEntryMappingTx(ctx, tx, statementID, cardEntryID); err != nil {
		return uuid.Nil, err
	}

	if err := s.statementRepo.AddPurchaseToStatementTx(ctx, tx, statementID, request.Amount); err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}

	cache.InvalidateCreditCard(ctx, s.redis, card.CardID, userID, card.ClearingAccountId)

	return transactionID, nil
}

// findOrCreateStatementTx relies on the caller holding the card lock, because
// credit_card_statements has no unique constraint on the cycle.
func (s *Service) findOrCreateStatementTx(
	ctx context.Context,
	tx pgx.Tx,
	cardID uuid.UUID,
	cycle BillingCycle,
) (uuid.UUID, error) {

	statement, err := s.statementRepo.FindStatementByCycleTx(ctx, tx, cardID, cycle)
	if err == nil {
		return statement.ID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}

	return s.statementRepo.CreateStatementTx(ctx, tx, CreditCardStatement{
		CardID:        cardID,
		StartDate:     cycle.CycleStartDate,
		EndDate:       cycle.CycleEndDate,
		StatementDate: cycle.StatementDate,
		DueDate:       cycle.DueDate,
		TotalAmount:   decimal.Zero,
		Status:        StatementOpen,
	})
}
