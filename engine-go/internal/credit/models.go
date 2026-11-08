package credit

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type StatementStatus string

const (
	StatementOpen StatementStatus = "OPEN"
	StatementPaid StatementStatus = "PAID"
)

type CreditCard struct {
	CardID            uuid.UUID
	ClearingAccountId uuid.UUID
	CreditLimit       decimal.Decimal
	BillingDay        int
	PaymentDueDays    int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CreditCardStatement struct {
	ID            uuid.UUID
	CardID        uuid.UUID
	StartDate     time.Time
	EndDate       time.Time
	StatementDate time.Time
	DueDate       time.Time
	// TotalAmount is the amount still owed, except on a PAID statement,
	// which keeps the amount of its final payment.
	TotalAmount decimal.Decimal
	Status      StatementStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreditCardTransactionRequest struct {
	CardID       uuid.UUID
	Amount       decimal.Decimal
	Description  string
	PurchaseDate time.Time
}

type BillingCycle struct {
	CycleStartDate time.Time
	CycleEndDate   time.Time
	StatementDate  time.Time
	DueDate        time.Time
}

type CreditCardPaymentRequest struct {
	CardID           uuid.UUID
	StatementID      uuid.UUID
	PaymentAccountID uuid.UUID // cash/bank account
	Amount           decimal.Decimal
	Description      string
}
