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
	BNPL        AccountType = "BNPL"
	CLEARING    AccountType = "CLEARING"
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

type Account struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Type      string
	Balance   decimal.Decimal
	CreatedAt time.Time
}

type RecentTransaction struct {
	ID          uuid.UUID
	Description string
	Amount      decimal.Decimal
	Direction   string
	AccountName string
	CreatedAt   time.Time
}

type TransactionDetail struct {
	ID          uuid.UUID
	Description string
	CreatedAt   time.Time
	Entries     []TransactionEntry
}

type TransactionEntry struct {
	ID            uuid.UUID
	TransactionID uuid.UUID
	AccountID     uuid.UUID
	AccountName   string
	AccountType   AccountType
	Amount        decimal.Decimal
	Type          EntryType
}
