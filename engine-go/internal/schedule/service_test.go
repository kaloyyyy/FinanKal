package schedule

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
}

func TestPrincipal(t *testing.T) {
	d := decimal.RequireFromString
	tests := []struct {
		rate string
		k    int
		want string
	}{
		{"0.10", 1, "100"},
		{"0.10", 2, "90.91"},
		{"0", 3, "110"}, // no rate recorded: the whole installment
	}
	for _, test := range tests {
		if got := Principal(d("110"), d(test.rate), test.k); !got.Equal(d(test.want)) {
			t.Errorf("rate %s, installment %d: got %s, want %s", test.rate, test.k, got, test.want)
		}
	}
}

func TestGenerateDates(t *testing.T) {
	endDate := day(2026, 1, 15)
	dayOfMonth := func(d int) *int { return &d }

	tests := []struct {
		name      string
		scheduled ScheduledTransaction
		from, to  time.Time
		want      []time.Time
	}{
		{
			"once inside the window",
			ScheduledTransaction{Frequency: FrequencyOnce, StartDate: day(2026, 1, 10)},
			day(2026, 1, 1), day(2026, 1, 31),
			[]time.Time{day(2026, 1, 10)},
		},
		{
			"once before the window must not move to the window start",
			ScheduledTransaction{Frequency: FrequencyOnce, StartDate: day(2026, 1, 1)},
			day(2026, 2, 1), day(2026, 2, 28),
			nil,
		},
		{
			"weekly stays on the start weekday when the window starts later",
			ScheduledTransaction{Frequency: FrequencyWeekly, StartDate: day(2026, 1, 1)},
			day(2026, 1, 10), day(2026, 1, 31),
			[]time.Time{day(2026, 1, 15), day(2026, 1, 22), day(2026, 1, 29)},
		},
		{
			"daily stops at the end date",
			ScheduledTransaction{Frequency: FrequencyDaily, StartDate: day(2026, 1, 13), EndDate: &endDate},
			day(2026, 1, 1), day(2026, 1, 31),
			[]time.Time{day(2026, 1, 13), day(2026, 1, 14), day(2026, 1, 15)},
		},
		{
			"monthly clamps to short months and returns to the day",
			ScheduledTransaction{Frequency: FrequencyMonthly, StartDate: day(2026, 1, 1), DayOfMonth: dayOfMonth(31)},
			day(2026, 1, 1), day(2026, 3, 31),
			[]time.Time{day(2026, 1, 31), day(2026, 2, 28), day(2026, 3, 31)},
		},
		{
			"monthly skips a day before the start",
			ScheduledTransaction{Frequency: FrequencyMonthly, StartDate: day(2026, 1, 20), DayOfMonth: dayOfMonth(5)},
			day(2026, 1, 1), day(2026, 2, 28),
			[]time.Time{day(2026, 2, 5)},
		},
		{
			"yearly from Feb 29 returns to Feb 29",
			ScheduledTransaction{Frequency: FrequencyYearly, StartDate: day(2024, 2, 29)},
			day(2028, 1, 1), day(2028, 12, 31),
			[]time.Time{day(2028, 2, 29)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := generateDates(&test.scheduled, test.from, test.to)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestGenerateDatesRequiresDayOfMonth(t *testing.T) {
	_, err := generateDates(&ScheduledTransaction{Frequency: FrequencyMonthly}, day(2026, 1, 1), day(2026, 2, 1))
	if err == nil {
		t.Error("want an error")
	}
}

func TestGenerateOccurrencesSkipsInactiveSchedules(t *testing.T) {
	// A nil repository would panic if this reached the database.
	occurrences, err := (&Service{}).GenerateOccurrences(context.Background(), &ScheduledTransaction{
		Frequency: FrequencyDaily,
		StartDate: day(2026, 1, 1),
	}, day(2026, 1, 1), day(2026, 1, 31))

	if err != nil || occurrences != nil {
		t.Errorf("got %v, %v", occurrences, err)
	}
}

func TestValidateInstallments(t *testing.T) {
	src, loan := uuid.New(), uuid.New()
	pay := func(d time.Time, amount int64) Installment { return Installment{DueDate: d, Amount: decimal.NewFromInt(amount)} }
	ok := []Installment{pay(day(2026, 12, 4), 900), pay(day(2027, 1, 4), 900)}

	tests := []struct {
		name         string
		loanName     string
		source       uuid.UUID
		installments []Installment
		wantErr      bool
	}{
		{"valid schedule", "Mari Loan", src, ok, false},
		{"blank name", "  ", src, ok, true},
		{"paying the loan from itself", "Mari Loan", loan, ok, true},
		{"no installments", "Mari Loan", src, nil, true},
		{"zero amount", "Mari Loan", src, []Installment{pay(day(2026, 12, 4), 0)}, true},
		{"same due date twice", "Mari Loan", src, []Installment{ok[0], ok[0]}, true},
		{"dates out of order", "Mari Loan", src, []Installment{ok[1], ok[0]}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInstallments(strings.TrimSpace(tt.loanName), tt.source, loan, tt.installments)
			if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, ErrInvalidSchedule)) {
				t.Fatalf("validateInstallments() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
