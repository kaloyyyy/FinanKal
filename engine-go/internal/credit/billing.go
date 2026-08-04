package credit

import "time"

// CalculateBillingCycle determines the billing cycle for a purchase.
func CalculateBillingCycle(
	purchaseDate time.Time,
	card CreditCard,
) BillingCycle {

	statementDate := CalculateStatementDate(purchaseDate, card)
	startDate := CalculateStartDate(statementDate)
	endDate := statementDate
	dueDate := CalculateDueDate(statementDate, card)

	return BillingCycle{
		CycleStartDate: startDate,
		CycleEndDate:   endDate,
		StatementDate:  statementDate,
		DueDate:        dueDate,
	}
}

// CalculateStatementDate returns the statement date that a purchase belongs to.
func CalculateStatementDate(
	purchaseDate time.Time,
	card CreditCard,
) time.Time {

	location := purchaseDate.Location()

	if purchaseDate.Day() <= card.BillingDay {
		return time.Date(
			purchaseDate.Year(),
			purchaseDate.Month(),
			card.BillingDay,
			0, 0, 0, 0,
			location,
		)
	}

	nextMonth := purchaseDate.AddDate(0, 1, 0)

	return time.Date(
		nextMonth.Year(),
		nextMonth.Month(),
		card.BillingDay,
		0, 0, 0, 0,
		location,
	)
}

// CalculateStartDate returns the first day of the billing cycle.
func CalculateStartDate(
	statementDate time.Time,
) time.Time {

	previousStatement := statementDate.AddDate(0, -1, 0)

	return previousStatement.AddDate(0, 0, 1)
}

// CalculateDueDate returns the payment due date.
func CalculateDueDate(
	statementDate time.Time,
	card CreditCard,
) time.Time {

	return statementDate.AddDate(0, 0, card.PaymentDueDays)
}

// IsWithinBillingCycle checks if a purchase belongs to the cycle.
func IsWithinBillingCycle(
	purchaseDate time.Time,
	cycle BillingCycle,
) bool {

	return !purchaseDate.Before(cycle.CycleStartDate) &&
		!purchaseDate.After(cycle.CycleEndDate)
}
