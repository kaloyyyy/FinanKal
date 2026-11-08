package credit

import (
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestCalculateBillingCycle(t *testing.T) {
	card := CreditCard{BillingDay: 15, PaymentDueDays: 20}

	tests := []struct {
		name                 string
		purchase, start, end time.Time
	}{
		{"before billing day", date(2026, 3, 10), date(2026, 2, 16), date(2026, 3, 15)},
		{"on billing day", date(2026, 3, 15), date(2026, 2, 16), date(2026, 3, 15)},
		{"after billing day", date(2026, 3, 16), date(2026, 3, 16), date(2026, 4, 15)},
		{"month end must not skip February", date(2026, 1, 31), date(2026, 1, 16), date(2026, 2, 15)},
		{"crosses the year", date(2026, 12, 20), date(2026, 12, 16), date(2027, 1, 15)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cycle := CalculateBillingCycle(test.purchase, card)

			if !cycle.CycleStartDate.Equal(test.start) || !cycle.CycleEndDate.Equal(test.end) {
				t.Errorf("got %s to %s, want %s to %s", cycle.CycleStartDate, cycle.CycleEndDate, test.start, test.end)
			}
			if want := test.end.AddDate(0, 0, 20); !cycle.DueDate.Equal(want) {
				t.Errorf("due date %s, want %s", cycle.DueDate, want)
			}
		})
	}
}
