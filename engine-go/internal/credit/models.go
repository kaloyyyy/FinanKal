package credit

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type StatementStatus string

const (
	StatementOpen   StatementStatus = "OPEN"
	StatementClosed StatementStatus = "CLOSED"
	StatementPaid   StatementStatus = "PAID"
)

type CreditCard struct {
	ID             uuid.UUID
	AccountID      uuid.UUID
	CreditLimit    decimal.Decimal
	BillingDay     int
	PaymentDueDays int
	CutoffTime     time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreditCardStatement struct {
	ID            uuid.UUID
	CreditCardID  uuid.UUID
	StartDate     time.Time
	EndDate       time.Time
	StatementDate time.Time
	DueDate       time.Time
	TotalAmount   decimal.Decimal
	Status        StatementStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreditCardEntry struct {
	ID          uuid.UUID
	StatementID uuid.UUID
	EntryID     uuid.UUID
	CreatedAt   time.Time
}

type CreditCardTransaction struct {
	TransactionID    uuid.UUID
	CreditCardID     uuid.UUID
	ExpenseAccountID uuid.UUID
	Amount           decimal.Decimal
	Description      string
	TransactionDate  time.Time
}

type BillingCycle struct {
	CycleStartDate time.Time
	CycleEndDate   time.Time
	StatementDate  time.Time
	DueDate        time.Time
}
