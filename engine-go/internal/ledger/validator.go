package ledger

import (
	"errors"
	"fmt"

	"github.com/kaloy/finankal/engine-go/internal/model"
	"github.com/shopspring/decimal"
)

func validateEntries(entries []model.Entry) error {
	debit := decimal.Zero
	credit := decimal.Zero

	for _, e := range entries {
		fmt.Println("Validating entry:", e.AccountID, e.Amount, e.Type)
		switch e.Type {
		case model.DEBIT:
			debit = debit.Add(e.Amount)
		case model.CREDIT:
			credit = credit.Add(e.Amount)
		default:
			return errors.New("invalid entry type")
		}
	}

	if !debit.Equal(credit) {
		return errors.New("entries are not balanced")
	}

	return nil
}
