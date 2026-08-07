package credit

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/db"
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
	log.Default().Println("Recording credit card transaction...")
	log.Default().Println("Request:", request)
	// Get credit card
	card, err := s.cardRepo.GetCreditCard(
		ctx,
		request.CardID,
	)

	log.Default().Println("Credit card:", card)
	if err != nil {
		db.LogPgError(err)
		return uuid.Nil, err
	}

	// Validate transaction
	err = s.ValidateCreditCardTransaction(
		ctx,
		request,
		*card,
	)

	if err != nil {
		db.LogPgError(err)
		return uuid.Nil, err
	}

	// Begin database transaction
	tx, err := s.ledgerRepo.BeginTx(ctx)

	if err != nil {
		db.LogPgError(err)
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
					CardID:        request.CardID,
					StartDate:     cycle.CycleStartDate,
					EndDate:       cycle.CycleEndDate,
					StatementDate: cycle.StatementDate,
					DueDate:       cycle.DueDate,
					TotalAmount:   decimal.Zero,
					Status:        StatementOpen,
				},
			)
		if err != nil {
			log.Printf("CreateStatementTx error: %#v", err)
			log.Printf("CreateStatementTx error: %v", err)
			return uuid.Nil, err
		}

		statement, err =
			s.statementRepo.GetStatementTx(
				ctx,
				tx,
				statementID,
			)
		log.Default().Println("Created new statement:", statement)
		if err != nil {
			db.LogPgError(err)
			return uuid.Nil, err
		}
	}

	// Get owner
	userID, err :=
		s.ledgerRepo.GetUserIDFromAccount(
			ctx,
			card.CardID,
		)

	if err != nil {
		db.LogPgError(err)
		return uuid.Nil, err
	}
	log.Default().Println("User:", userID)
	// Create ledger transaction
	transactionID, err :=
		s.ledgerRepo.CreateTransactionTx(
			ctx,
			tx,
			userID,
			request.Description,
		)

	if err != nil {
		log.Default().Println("Error creating transaction:", err)
		db.LogPgError(err)
		return uuid.Nil, err
	}

	// Debit expense
	log.Default().Println("Creating debit entry...")
	_, err =
		s.ledgerRepo.InsertEntryTx(
			ctx,
			tx,
			transactionID,
			card.ClearingAccountId,
			request.Amount,
			ledger.DEBIT,
		)

	if err != nil {
		db.LogPgError(err)
		return uuid.Nil, err
	}

	// Credit credit card liability
	cardEntryID, err := s.ledgerRepo.InsertEntryTx(ctx, tx, transactionID, card.CardID, request.Amount,
		ledger.CREDIT)

	if err != nil {
		db.LogPgError(err)
		return uuid.Nil, err
	}

	// Attach entry to statement
	log.Default().Println("Creating entry mapping...")
	err = s.entryRepo.CreateEntryMappingTx(
		ctx,
		tx,
		statement.ID,
		cardEntryID,
	)

	if err != nil {
		db.LogPgError(err)
		return uuid.Nil, err
	}

	log.Default().Println("Incrementing Statement Total...")
	// Update statement total
	err =
		s.statementRepo.IncrementStatementTotalTx(
			ctx,
			tx,
			statement.ID,
			request.Amount,
		)

	if err != nil {
		db.LogPgError(err)
		return uuid.Nil, err
	}

	log.Default().Println("Committing transaction...")
	// Commit transaction
	err = tx.Commit(ctx)

	if err != nil {
		db.LogPgError(err)
		return uuid.Nil, err
	}

	// Redis invalidation
	_ = cache.InvalidateTransaction(
		ctx,
		s.redis,
		userID,
		card.CardID,
		card.ClearingAccountId,
	)

	// Credit card specific cache
	_ = cache.InvalidateCreditCard(
		ctx,
		s.redis,
		request.CardID,
		userID,
	)

	return transactionID, nil
}
