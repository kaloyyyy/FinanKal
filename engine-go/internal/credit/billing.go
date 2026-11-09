package credit

import "time"

// CalculateBillingCycle determines the billing cycle for a purchase.
func CalculateBillingCycle(purchaseDate time.Time, card CreditCard) BillingCycle {
	statementDate := CalculateStatementDate(purchaseDate, card)

	return BillingCycle{
		CycleStartDate: CalculateStartDate(statementDate, card),
		CycleEndDate:   statementDate,
		StatementDate:  statementDate,
		DueDate:        CalculateDueDate(statementDate, card),
	}
}

// StatementDateIn returns the statement date of the cycle that closes in the given month: the card's
// billing day, moved off a weekend by its StatementShift (unset = stays on the day).
func StatementDateIn(year int, month time.Month, card CreditCard, loc *time.Location) time.Time {
	// time.Date normalises month 13 into January of the next year, and so on.
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	day := card.BillingDay
	if last := first.AddDate(0, 1, -1).Day(); day > last {
		day = last
	}
	billing := time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, loc)

	switch card.StatementShift {
	case ShiftNextBusinessDay:
		for isWeekend(billing) {
			billing = billing.AddDate(0, 0, 1)
		}
	case ShiftNearestBusinessDay:
		switch billing.Weekday() {
		case time.Saturday:
			billing = billing.AddDate(0, 0, -1)
		case time.Sunday:
			billing = billing.AddDate(0, 0, 1)
		}
	}
	return billing
}

// CalculateStatementDate returns the statement date that a purchase belongs to: this month's
// statement if the purchase is on or before it, otherwise next month's.
func CalculateStatementDate(purchaseDate time.Time, card CreditCard) time.Time {
	statement := StatementDateIn(purchaseDate.Year(), purchaseDate.Month(), card, purchaseDate.Location())
	if purchaseDate.After(statement) {
		// Month + 1, not AddDate(0, 1, 0): that turns Jan 31 into Mar 3 and skips February's statement.
		statement = StatementDateIn(purchaseDate.Year(), purchaseDate.Month()+1, card, purchaseDate.Location())
	}
	return statement
}

// NextStatementDate returns the first statement date strictly after the given day.
func NextStatementDate(after time.Time, card CreditCard) time.Time {
	statement := CalculateStatementDate(after, card)
	if !statement.After(after) {
		statement = StatementDateIn(after.Year(), after.Month()+1, card, after.Location())
	}
	return statement
}

// CalculateStartDate returns the first day of the billing cycle: the day after the previous statement.
func CalculateStartDate(statementDate time.Time, card CreditCard) time.Time {
	previous := StatementDateIn(statementDate.Year(), statementDate.Month()-1, card, statementDate.Location())
	return previous.AddDate(0, 0, 1)
}

func isWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

// CalculateDueDate returns the payment due date, following the card's rule (DAYS when unset):
//   - DAYS: PaymentDueDays calendar days after the statement date.
//   - DAY_OF_MONTH: the next PaymentDueDays-th of a month after the statement date (the 5th, say).
//   - BUSINESS_DAYS: PaymentDueDays weekdays after the statement date.
//
// shortcut: business days skip weekends only, not public holidays; add a holiday calendar if a due date lands on one.
func CalculateDueDate(statementDate time.Time, card CreditCard) time.Time {
	switch card.DueRule {
	case DueDayOfMonth:
		day := card.PaymentDueDays
		next := firstOfMonth(statementDate)
		if statementDate.Day() >= day {
			next = next.AddDate(0, 1, 0)
		}
		if last := next.AddDate(0, 1, -1).Day(); day > last {
			day = last // the 31st in a 30-day month is its last day
		}
		return time.Date(next.Year(), next.Month(), day, 0, 0, 0, 0, statementDate.Location())
	case DueBusinessDays:
		due := statementDate
		for left := card.PaymentDueDays; left > 0; {
			due = due.AddDate(0, 0, 1)
			if !isWeekend(due) {
				left--
			}
		}
		return due
	default:
		return statementDate.AddDate(0, 0, card.PaymentDueDays)
	}
}

func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}
