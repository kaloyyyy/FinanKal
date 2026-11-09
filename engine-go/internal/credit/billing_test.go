package credit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

func TestCalculateDueDateRules(t *testing.T) {
	tests := []struct {
		name      string
		card      CreditCard
		statement time.Time
		want      time.Time
	}{
		{"calendar days (default rule)", CreditCard{PaymentDueDays: 14}, date(2026, 10, 23), date(2026, 11, 6)},
		{"day of month, later this month", CreditCard{DueRule: DueDayOfMonth, PaymentDueDays: 28}, date(2026, 10, 24), date(2026, 10, 28)},
		{"day of month, next month (Atome: 5th)", CreditCard{DueRule: DueDayOfMonth, PaymentDueDays: 5}, date(2026, 11, 24), date(2026, 12, 5)},
		{"day of month on the statement day is next month", CreditCard{DueRule: DueDayOfMonth, PaymentDueDays: 24}, date(2026, 11, 24), date(2026, 12, 24)},
		{"day of month across the year", CreditCard{DueRule: DueDayOfMonth, PaymentDueDays: 5}, date(2026, 12, 24), date(2027, 1, 5)},
		// Fri 23 Oct + 11 weekdays = Mon 9 Nov (UnionBank's printed due date).
		{"business days skip weekends", CreditCard{DueRule: DueBusinessDays, PaymentDueDays: 11}, date(2026, 10, 23), date(2026, 11, 9)},
		{"business days from a Saturday", CreditCard{DueRule: DueBusinessDays, PaymentDueDays: 1}, date(2026, 10, 24), date(2026, 10, 26)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CalculateDueDate(test.statement, test.card); !got.Equal(test.want) {
				t.Errorf("due %s, want %s", got.Format("Mon 2 Jan 2006"), test.want.Format("Mon 2 Jan 2006"))
			}
		})
	}
}

func TestAllocatePayments(t *testing.T) {
	card, other := uuid.New(), uuid.New()
	d := decimal.RequireFromString
	bills := []Bill{
		{CardID: card, StatementDate: date(2026, 10, 23), Billed: d("21247.65")}, // given out of order on purpose
		{CardID: card, StatementDate: date(2026, 9, 23), Billed: d("14485.15")},
		{CardID: other, StatementDate: date(2026, 10, 24), Billed: d("19326.83")},
	}
	payments := []CardPayment{
		{CardID: card, Date: date(2026, 9, 10), Amount: d("14057.15")},  // before any tracked bill: an older balance
		{CardID: card, Date: date(2026, 9, 25), Amount: d("14485.15")},  // pays the Sep bill
		{CardID: card, Date: date(2026, 10, 23), Amount: d("100.00")},   // on the Oct statement date: already on it, pays the Sep bill (already paid) so it carries nowhere
		{CardID: card, Date: date(2026, 11, 3), Amount: d("18912.67")},  // part of the Oct bill
		{CardID: other, Date: date(2026, 11, 4), Amount: d("19326.83")}, // another card's bill: no effect here
	}

	got := AllocatePayments(bills, payments)
	owed := map[string]string{}
	for _, b := range got {
		owed[b.CardID.String()[:4]+b.StatementDate.Format(" Jan")] = b.Owed.StringFixed(2)
	}
	want := map[string]string{
		card.String()[:4] + " Sep":  "0.00",
		card.String()[:4] + " Oct":  "2334.98", // 21,247.65 − 18,912.67
		other.String()[:4] + " Oct": "0.00",
	}
	for k, v := range want {
		if owed[k] != v {
			t.Errorf("%s owed %s, want %s (all: %v)", k, owed[k], v, owed)
		}
	}
}

func TestCurrentCycles(t *testing.T) {
	d := decimal.RequireFromString
	ub := CreditCard{CardID: uuid.New(), BillingDay: 23, DueRule: DueBusinessDays, PaymentDueDays: 11}
	atome := CreditCard{CardID: uuid.New(), BillingDay: 24, DueRule: DueDayOfMonth, PaymentDueDays: 7}
	all := []Bill{
		{CardID: ub.CardID, StatementDate: date(2026, 10, 23)},
		{CardID: atome.CardID, StatementDate: date(2026, 11, 24)},
	}
	open := []Bill{{CardID: atome.CardID, StatementDate: date(2026, 11, 24), Owed: d("22111.07")}} // Atome's open bill already covers it
	cards := []CardOwing{{Card: ub, Name: "UB", Owed: d("20824.75")}, {Card: atome, Name: "Atome", Owed: d("22111.07")}}

	got := CurrentCycles(cards, open, all, date(2026, 11, 9))
	if len(got) != 1 {
		t.Fatalf("got %d cycles, want only UnionBank's (Atome's bill already covers what it owes): %+v", len(got), got)
	}
	c := got[0]
	// Next statement 23 Nov (after the 23 Oct bill and today), due 11 weekdays later: Tue 8 Dec.
	if !c.StatementDate.Equal(date(2026, 11, 23)) || !c.DueDate.Equal(date(2026, 12, 8)) || !c.Owed.Equal(d("20824.75")) || !c.SoFar {
		t.Errorf("got statement %s due %s owed %s so far %v; want 23 Nov, 8 Dec, 20824.75, true",
			c.StatementDate.Format("2 Jan"), c.DueDate.Format("2 Jan"), c.Owed, c.SoFar)
	}
	if !c.StartDate.Equal(date(2026, 10, 24)) {
		t.Errorf("cycle starts %s, want 24 Oct", c.StartDate.Format("2 Jan"))
	}
}

func TestStatementDateMovesOffWeekends(t *testing.T) {
	card := func(shift string) CreditCard { return CreditCard{BillingDay: 23, StatementShift: shift} }
	tests := []struct {
		name  string
		card  CreditCard
		month time.Month
		year  int
		want  time.Time
	}{
		{"weekday stays put (Mon 23 Nov 2026)", card(ShiftNextBusinessDay), time.November, 2026, date(2026, 11, 23)},
		{"Saturday to next business day", card(ShiftNextBusinessDay), time.January, 2027, date(2027, 1, 25)},
		{"Saturday to nearest business day (Friday)", card(ShiftNearestBusinessDay), time.January, 2027, date(2027, 1, 22)},
		{"Saturday, no shift", card(ShiftNone), time.January, 2027, date(2027, 1, 23)},
		{"Sunday to next business day", card(ShiftNextBusinessDay), time.May, 2027, date(2027, 5, 24)},
		{"Sunday to nearest business day (Monday)", card(ShiftNearestBusinessDay), time.May, 2027, date(2027, 5, 24)},
		{"month 13 is next January", card(ShiftNone), 13, 2026, date(2027, 1, 23)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := StatementDateIn(test.year, test.month, test.card, time.UTC); !got.Equal(test.want) {
				t.Errorf("statement %s, want %s", got.Format("Mon 2 Jan 2006"), test.want.Format("Mon 2 Jan 2006"))
			}
		})
	}
}

func TestNextStatementDateIsThisMonthOrNext(t *testing.T) {
	ub := CreditCard{BillingDay: 23, StatementShift: ShiftNextBusinessDay}
	tests := []struct {
		name        string
		after, want time.Time
	}{
		{"before this month's statement: this month", date(2026, 11, 9), date(2026, 11, 23)},
		{"on the statement day: next month", date(2026, 11, 23), date(2026, 12, 23)},
		// 23 Jan 2027 is a Saturday, so January's statement is Mon 25 Jan: still ahead on the 23rd and 24th.
		{"on a weekend billing day: the shifted date this month", date(2027, 1, 23), date(2027, 1, 25)},
		{"after the shifted date: next month", date(2027, 1, 25), date(2027, 2, 23)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NextStatementDate(test.after, ub); !got.Equal(test.want) {
				t.Errorf("next statement %s, want %s", got.Format("Mon 2 Jan 2006"), test.want.Format("Mon 2 Jan 2006"))
			}
		})
	}
	// A cycle starts the day after the previous (shifted) statement.
	if got := CalculateStartDate(date(2027, 2, 23), ub); !got.Equal(date(2027, 1, 26)) {
		t.Errorf("February cycle starts %s, want Tue 26 Jan 2027", got.Format("Mon 2 Jan 2006"))
	}
}
