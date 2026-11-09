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

// How a card's due date follows its statement date (see CalculateDueDate).
const (
	DueDays         = "DAYS"
	DueDayOfMonth   = "DAY_OF_MONTH"
	DueBusinessDays = "BUSINESS_DAYS"
)

// Where a statement date goes when the billing day is a weekend (see StatementDateIn). Unset = NONE.
const (
	ShiftNone               = "NONE"
	ShiftNextBusinessDay    = "NEXT_BUSINESS_DAY"
	ShiftNearestBusinessDay = "NEAREST_BUSINESS_DAY"
)

// StatementEstimated marks a bill the engine worked out before its statement arrived; importing the
// statement replaces it.
const StatementEstimated StatementStatus = "ESTIMATED"

type CreditCard struct {
	CardID            uuid.UUID
	ClearingAccountId uuid.UUID
	CreditLimit       decimal.Decimal
	BillingDay        int
	PaymentDueDays    int
	DueRule           string // DueDays, DueDayOfMonth or DueBusinessDays
	StatementShift    string // ShiftNone, ShiftNextBusinessDay or ShiftNearestBusinessDay
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
