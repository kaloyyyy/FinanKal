package ledger

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func entry(entryType EntryType, amount string) Entry {
	return Entry{Type: entryType, Amount: decimal.RequireFromString(amount)}
}

func TestValidateEntries(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		valid   bool
	}{
		{"balanced", []Entry{entry(DEBIT, "10.50"), entry(CREDIT, "10.5")}, true},
		{"split", []Entry{entry(DEBIT, "10"), entry(CREDIT, "4"), entry(CREDIT, "6")}, true},
		{"trailing zeros are not extra precision", []Entry{entry(DEBIT, "1.500000"), entry(CREDIT, "1.5")}, true},
		{"no entries", nil, false},
		{"unbalanced", []Entry{entry(DEBIT, "10"), entry(CREDIT, "9")}, false},
		{"negative pair balances but moves money backwards", []Entry{entry(DEBIT, "-10"), entry(CREDIT, "-10")}, false},
		{"zero", []Entry{entry(DEBIT, "0"), entry(CREDIT, "0")}, false},
		{"rounds unbalanced in the database", []Entry{entry(DEBIT, "1.00005"), entry(CREDIT, "0.50003"), entry(CREDIT, "0.50002")}, false},
		{"unknown type", []Entry{entry("REFUND", "10"), entry(CREDIT, "10")}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateEntries(test.entries)

			if test.valid && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !test.valid && !errors.Is(err, ErrInvalidTransaction) {
				t.Errorf("got %v, want ErrInvalidTransaction", err)
			}
		})
	}
}
