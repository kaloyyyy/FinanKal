package loan

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

var mayaRate = decimal.RequireFromString("0.00133")

func amounts(cs []Charge) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.Due.Format("2 Jan") + " " + c.Amount.StringFixed(2)
	}
	return out
}

func assertCharges(t *testing.T, got []Charge, want ...string) {
	t.Helper()
	g := amounts(got)
	if len(g) != len(want) {
		t.Fatalf("got %v, want %v", g, want)
	}
	for i := range want {
		if g[i] != want[i] {
			t.Fatalf("got %v, want %v", g, want)
		}
	}
}

func TestChargesMatchMayaTableWhenPaidOnSchedule(t *testing.T) {
	// Rows 1–3 of the Maya Personal Loan disclosure: ₱120,000 at 0.133% a day, ₱7,862.36 a month.
	pay := decimal.RequireFromString("7862.36")
	payments := []Payment{{day(2026, 5, 1), pay}, {day(2026, 6, 1), pay}}
	dues := []time.Time{day(2026, 5, 1), day(2026, 6, 1), day(2026, 7, 1)}
	got := Charges(decimal.NewFromInt(120000), day(2026, 4, 1), dues, payments, mayaRate, day(2026, 7, 1))
	assertCharges(t, got, "1 May 4788.00", "1 Jun 4820.84", "1 Jul 4543.98")
}

func TestAPaymentBetweenDueDatesPaysInterestSoFarThenPrincipal(t *testing.T) {
	dues := []time.Time{day(2026, 5, 1)}
	payments := []Payment{{day(2026, 4, 11), decimal.NewFromInt(10000)}}
	got := Charges(decimal.NewFromInt(120000), day(2026, 4, 1), dues, payments, mayaRate, day(2026, 5, 1))
	// 10 days on 120,000 = 1,596.00; the other 8,404 is principal, so 20 days on 111,596 = 2,968.45.
	assertCharges(t, got, "11 Apr 1596.00", "1 May 2968.45")
}

func TestInterestSoFarIsChargedToday(t *testing.T) {
	// Maya app on 9 Nov 2026: remaining ₱96,938.33 = principal ₱94,426.59 + ₱2,511.74 interest since 20 Oct.
	dues := []time.Time{day(2026, 11, 1), day(2026, 12, 1)}
	got := Charges(decimal.RequireFromString("94426.59"), day(2026, 10, 20), dues, nil, mayaRate, day(2026, 11, 9))
	assertCharges(t, got, "1 Nov 1507.05", "9 Nov 1004.70")
}

func TestFromMayaAppFigures(t *testing.T) {
	// Principal ₱94,426.59 after the 20 Oct 2026 payments; nothing paid since.
	dues := []time.Time{day(2026, 11, 1), day(2026, 12, 1)}
	got := Charges(decimal.RequireFromString("94426.59"), day(2026, 10, 20), dues, nil, mayaRate, day(2026, 12, 1))
	// ₱125.5873647 a day: 12 days to 1 Nov, then 30 days to 1 Dec on the same principal.
	assertCharges(t, got, "1 Nov 1507.05", "1 Dec 3767.62")
}
