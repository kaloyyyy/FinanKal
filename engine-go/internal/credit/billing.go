package credit

import "time"

// CalculateBillingCycle determines the billing cycle for a purchase.
func CalculateBillingCycle(purchaseDate time.Time, card CreditCard) BillingCycle {
	statementDate := CalculateStatementDate(purchaseDate, card)

	return BillingCycle{
		CycleStartDate: CalculateStartDate(statementDate),
		CycleEndDate:   statementDate,
		StatementDate:  statementDate,
		DueDate:        CalculateDueDate(statementDate, card),
	}
}

// CalculateStatementDate returns the statement date that a purchase belongs to.
func CalculateStatementDate(purchaseDate time.Time, card CreditCard) time.Time {
	month := purchaseDate.Month()
	if purchaseDate.Day() > card.BillingDay {
		// Not AddDate(0, 1, 0): that turns Jan 31 into Mar 3 and skips February's statement.
		month++
	}

	return time.Date(purchaseDate.Year(), month, card.BillingDay, 0, 0, 0, 0, purchaseDate.Location())
}

// CalculateStartDate returns the first day of the billing cycle.
func CalculateStartDate(statementDate time.Time) time.Time {
	return statementDate.AddDate(0, -1, 1)
}

// CalculateDueDate returns the payment due date.
func CalculateDueDate(statementDate time.Time, card CreditCard) time.Time {
	return statementDate.AddDate(0, 0, card.PaymentDueDays)
}
