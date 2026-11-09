// Package loan handles loans whose interest is charged on what's still owed (Maya Personal Loan).
// Unlike loans booked with all their interest up front, paying ahead lowers the principal and with it
// every later interest charge, so interest is booked as it builds up, from the ledger's payments.
package loan

import (
	"time"

	"github.com/shopspring/decimal"
)

// Payment is money that reached the loan on a calendar date (UTC midnight).
type Payment struct {
	Date   time.Time
	Amount decimal.Decimal
}

// Charge is interest booked on a day: a due date, or a payment day (the interest that payment settles first).
type Charge struct {
	Due    time.Time
	Amount decimal.Decimal
}

// Charges works out the interest booked after from (the end of the day principal was last known, with no
// interest owed) up to today. Interest accrues daily at rate on the principal, counting from's own day. On a
// due date, a payment day or today the interest built up so far is charged; a payment pays charged interest first
// and the rest lowers the principal from that day on. Payments on or before from are already in principal.
// This is Maya Personal Loan's rule: paid on schedule it reproduces the amortization table, and from the
// app's principal after the 20 Oct 2026 payments it gives the app's 9 Nov interest (₱2,511.74) to a centavo.
func Charges(principal decimal.Decimal, from time.Time, dues []time.Time, payments []Payment, rate decimal.Decimal, today time.Time) []Charge {
	isDue := map[time.Time]bool{}
	for _, d := range dues {
		isDue[d] = true
	}
	var out []Charge
	p, unpaid, accrued := principal, decimal.Zero, decimal.Zero
	next := 0
	for day := from; !day.After(today); day = day.AddDate(0, 0, 1) {
		var paid decimal.Decimal
		for ; next < len(payments) && !payments[next].Date.After(day); next++ {
			if payments[next].Date.After(from) {
				paid = paid.Add(payments[next].Amount)
			}
		}
		// Today too, so what's owed matches the lender's "remaining" (principal + interest so far).
		if day.After(from) && (isDue[day] || paid.IsPositive() || day.Equal(today)) {
			if charge := accrued.Round(2); charge.IsPositive() {
				out = append(out, Charge{Due: day, Amount: charge})
				unpaid = unpaid.Add(charge)
			}
			accrued = decimal.Zero
		}
		toInterest := decimal.Min(paid, unpaid)
		unpaid = unpaid.Sub(toInterest)
		p = p.Sub(paid.Sub(toInterest))
		if day.Before(today) && p.IsPositive() {
			accrued = accrued.Add(p.Mul(rate))
		}
	}
	return out
}
