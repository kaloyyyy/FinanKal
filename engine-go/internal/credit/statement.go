package credit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// GetOrCreateStatement returns the statement for the billing cycle,
// creating one if it does not already exist.
func (s *Service) GetOrCreateStatement(
	ctx context.Context,
	card CreditCard,
	purchaseDate time.Time,
) (*CreditCardStatement, error) {

	cycle := CalculateBillingCycle(purchaseDate, card)

	statement, err := s.statementRepo.FindStatementByCycle(
		ctx,
		card.ID,
		cycle,
	)
	if err == nil {
		return statement, nil
	}

	if err != pgx.ErrNoRows {
		return nil, err
	}

	return s.CreateStatement(
		ctx,
		card,
		cycle,
	)
}

// CreateStatement creates a new billing statement.
func (s *Service) CreateStatement(
	ctx context.Context,
	card CreditCard,
	cycle BillingCycle,
) (*CreditCardStatement, error) {

	statement := CreditCardStatement{
		CreditCardID:  card.ID,
		StartDate:     cycle.CycleStartDate,
		EndDate:       cycle.CycleEndDate,
		StatementDate: cycle.StatementDate,
		DueDate:       cycle.DueDate,
		TotalAmount:   decimal.Zero,
		Status:        StatementOpen,
	}

	id, err := s.statementRepo.CreateStatement(
		ctx,
		statement,
	)
	if err != nil {
		return nil, err
	}

	statement.ID = id

	return &statement, nil
}

// UpdateStatementTotal updates the statement balance.
func (s *Service) UpdateStatementTotal(
	ctx context.Context,
	statementID uuid.UUID,
	total decimal.Decimal,
) error {

	return s.statementRepo.UpdateStatementTotal(
		ctx,
		statementID,
		total,
	)
}

// CloseStatement marks a statement as CLOSED.
func (s *Service) CloseStatement(
	ctx context.Context,
	statementID uuid.UUID,
) error {

	return s.statementRepo.UpdateStatementStatus(
		ctx,
		statementID,
		StatementClosed,
	)
}

// MarkStatementPaid marks a statement as PAID.
func (s *Service) MarkStatementPaid(
	ctx context.Context,
	statementID uuid.UUID,
) error {

	return s.statementRepo.UpdateStatementStatus(
		ctx,
		statementID,
		StatementPaid,
	)
}

func (s *Service) IncrementStatementTotal(
	ctx context.Context,
	statementID uuid.UUID,
	amount decimal.Decimal,
) error {

	return s.statementRepo.IncrementStatementTotal(
		ctx,
		statementID,
		amount,
	)
}
