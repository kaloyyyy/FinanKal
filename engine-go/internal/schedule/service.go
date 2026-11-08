package schedule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	scheduleRepo *ScheduleRepository
	ledgerRepo   *ledger.LedgerRepository
	redis        *redis.Client
}

func NewService(scheduleRepo *ScheduleRepository, ledgerRepo *ledger.LedgerRepository, redis *redis.Client) *Service {
	return &Service{scheduleRepo: scheduleRepo, ledgerRepo: ledgerRepo, redis: redis}
}

func (s *Service) GenerateOccurrences(
	ctx context.Context,
	scheduled *ScheduledTransaction,
	from time.Time,
	to time.Time,
) ([]ScheduledTransactionOccurrence, error) {

	if scheduled == nil {
		return nil, errors.New("scheduled transaction is nil")
	}

	if from.After(to) {
		return nil, errors.New("from date must not be after to date")
	}

	if !scheduled.Active {
		return nil, nil
	}

	dates, err := generateDates(scheduled, from, to)
	if err != nil {
		return nil, err
	}

	var occurrences []ScheduledTransactionOccurrence

	for _, date := range dates {
		occurrence, err := s.scheduleRepo.CreateOccurrence(ctx, scheduled.ID, date, scheduled.Amount, nil)
		if err != nil {
			return nil, err
		}

		// Already exists.
		if occurrence == nil {
			continue
		}

		occurrences = append(occurrences, *occurrence)
	}

	return occurrences, nil
}

// generateDates returns the schedule's dates inside [from, to]. Dates always count from
// StartDate, so overlapping windows produce the same dates and CreateOccurrence can dedupe them.
func generateDates(
	scheduled *ScheduledTransaction,
	from time.Time,
	to time.Time,
) ([]time.Time, error) {

	if scheduled.EndDate != nil && scheduled.EndDate.Before(to) {
		to = *scheduled.EndDate
	}

	start := scheduled.StartDate

	switch scheduled.Frequency {
	case FrequencyOnce:
		if start.Before(from) || start.After(to) {
			return nil, nil
		}
		return []time.Time{start}, nil

	case FrequencyDaily:
		return datesInWindow(from, to, func(n int) time.Time { return start.AddDate(0, 0, n) }), nil

	case FrequencyWeekly:
		return datesInWindow(from, to, func(n int) time.Time { return start.AddDate(0, 0, 7*n) }), nil

	case FrequencyMonthly:
		if scheduled.DayOfMonth == nil {
			return nil, errors.New("day_of_month is required for monthly schedule")
		}
		return generateMonthlyDates(start, from, to, *scheduled.DayOfMonth), nil

	case FrequencyYearly:
		return datesInWindow(from, to, func(n int) time.Time { return start.AddDate(n, 0, 0) }), nil

	default:
		return nil, fmt.Errorf("unsupported frequency: %s", scheduled.Frequency)
	}
}

// datesInWindow walks nth(0), nth(1), ... in ascending order, keeping those in [from, to].
func datesInWindow(from, to time.Time, nth func(n int) time.Time) []time.Time {
	var dates []time.Time

	for n := 0; ; n++ {
		date := nth(n)
		if date.After(to) {
			return dates
		}
		if !date.Before(from) {
			dates = append(dates, date)
		}
	}
}

func generateMonthlyDates(start, from, to time.Time, dayOfMonth int) []time.Time {
	// The first month's date can fall before start, e.g. day 5 for a schedule starting on the 20th.
	if from.Before(start) {
		from = start
	}

	return datesInWindow(from, to, func(n int) time.Time {
		firstOfMonth := time.Date(start.Year(), start.Month()+time.Month(n), 1, 0, 0, 0, 0, start.Location())
		lastDay := firstOfMonth.AddDate(0, 1, -1).Day()

		return firstOfMonth.AddDate(0, 0, min(dayOfMonth, lastDay)-1)
	})
}
