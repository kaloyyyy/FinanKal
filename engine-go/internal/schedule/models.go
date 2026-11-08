package schedule

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	ScheduleTypeRecurring = "RECURRING"
	ScheduleTypeOnce      = "ONCE"
)

const (
	TransactionTypeIncome   = "INCOME"
	TransactionTypeExpense  = "EXPENSE"
	TransactionTypeTransfer = "TRANSFER"
)

const (
	FrequencyOnce    = "ONCE"
	FrequencyDaily   = "DAILY"
	FrequencyWeekly  = "WEEKLY"
	FrequencyMonthly = "MONTHLY"
	FrequencyYearly  = "YEARLY"
)

const (
	AmountTypeFixed    = "FIXED"
	AmountTypeVariable = "VARIABLE"
)

const (
	OccurrencePending   = "PENDING"
	OccurrenceGenerated = "GENERATED"
	OccurrenceCompleted = "COMPLETED"
	OccurrenceSkipped   = "SKIPPED"
)

type ScheduledTransaction struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	Name                 string
	ScheduleType         string
	TransactionType      string
	Amount               *decimal.Decimal
	SourceAccountID      *uuid.UUID
	DestinationAccountID *uuid.UUID
	CategoryAccountID    *uuid.UUID
	StartDate            time.Time
	EndDate              *time.Time
	Frequency            string
	DayOfMonth           *int
	Active               bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
	AmountType           string
}

type ScheduledTransactionOccurrence struct {
	ID                     uuid.UUID
	ScheduledTransactionID uuid.UUID
	TransactionID          *uuid.UUID
	ScheduledDate          time.Time
	InstallmentNumber      *int
	Amount                 *decimal.Decimal
	Status                 string
	CreatedAt              time.Time
}
