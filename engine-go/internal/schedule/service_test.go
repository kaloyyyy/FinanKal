package schedule

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
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
