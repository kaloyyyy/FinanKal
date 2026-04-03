package model

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
	Type      string
	Balance   decimal.Decimal
	CreatedAt time.Time
}
