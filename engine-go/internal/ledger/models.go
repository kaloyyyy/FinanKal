package ledger

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EntryType string

const (
	DEBIT  EntryType = "DEBIT"
	CREDIT EntryType = "CREDIT"
)

type AccountType string

const (
	ASSET       AccountType = "ASSET"
	LIABILITY   AccountType = "LIABILITY"
	EXPENSE     AccountType = "EXPENSE"
	INCOME      AccountType = "INCOME"
	EQUITY      AccountType = "EQUITY"
	CREDIT_CARD AccountType = "CREDIT_CARD"
)

type Entry struct {
	ID            uuid.UUID
	TransactionID uuid.UUID
	AccountID     uuid.UUID
	Amount        decimal.Decimal
	CreatedAt     time.Time
	Type          EntryType
	Description   string
}

type AccountSummary struct {
	AccountID uuid.UUID
	Name      string
	Type      AccountType
	Balance   decimal.Decimal
	CreatedAt time.Time
}

type User struct {
	ID        uuid.UUID
	Name      string
	Username  string
	CreatedAt time.Time
}
