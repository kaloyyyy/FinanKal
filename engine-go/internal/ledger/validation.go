package ledger

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// ErrInvalidTransaction is wrapped by every error caused by a bad transaction request.
var ErrInvalidTransaction = errors.New("invalid transaction")

func validateEntries(entries []Entry) error {
	if len(entries) == 0 {
		return fmt.Errorf("%w: no entries provided", ErrInvalidTransaction)
	}

	debit := decimal.Zero
	credit := decimal.Zero

	for _, e := range entries {
		if !e.Amount.IsPositive() {
			return fmt.Errorf("%w: entry amount must be greater than zero", ErrInvalidTransaction)
		}

		// entries.amount is numeric(19,4). Postgres would round anything finer,
		// which can leave a transaction that balances here unbalanced once stored.
		if !e.Amount.Equal(e.Amount.Round(4)) {
			return fmt.Errorf("%w: entry amount has more than 4 decimal places", ErrInvalidTransaction)
		}

		switch e.Type {
		case DEBIT:
			debit = debit.Add(e.Amount)
		case CREDIT:
			credit = credit.Add(e.Amount)
		default:
			return fmt.Errorf("%w: invalid entry type", ErrInvalidTransaction)
		}
	}

	if !debit.Equal(credit) {
		return fmt.Errorf("%w: entries are not balanced", ErrInvalidTransaction)
	}

	return nil
}
