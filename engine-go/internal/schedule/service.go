package schedule

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type Service struct {
	scheduleRepo *ScheduleRepository
	ledgerRepo   *ledger.LedgerRepository
	redis        *redis.Client
}

func NewService(scheduleRepo *ScheduleRepository, ledgerRepo *ledger.LedgerRepository, redis *redis.Client) *Service {
	return &Service{scheduleRepo: scheduleRepo, ledgerRepo: ledgerRepo, redis: redis}
}

// GetPayables returns the user's unpaid loan and card installments, soonest first.
func (s *Service) GetPayables(ctx context.Context, userID uuid.UUID) ([]Payable, error) {
	return s.scheduleRepo.GetPayables(ctx, userID)
}

// ErrInvalidSchedule is returned for an installment schedule that can't be recorded; the message says why.
var ErrInvalidSchedule = errors.New("invalid schedule")

// Installment is one payment of a loan: how much, and when it's due (a calendar day).
type Installment struct {
	DueDate time.Time
	Amount  decimal.Decimal
}

// CreateInstallments records a loan's repayment schedule, paid from sourceID into accountID, and returns its id.
// The installments are numbered from 1 in the order given.
func (s *Service) CreateInstallments(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	sourceID, accountID uuid.UUID,
	installments []Installment,
) (uuid.UUID, error) {
	name = strings.TrimSpace(name)
	if err := validateInstallments(name, sourceID, accountID, installments); err != nil {
		return uuid.Nil, err
	}
	return s.scheduleRepo.CreateInstallments(ctx, userID, name, sourceID, accountID, installments)
}

func (s *Service) UpdateInstallments(ctx context.Context, userID, scheduleID, sourceID uuid.UUID, name string, firstNumber int, installments []Installment) error {
	name = strings.TrimSpace(name)
	if firstNumber < 1 || firstNumber+len(installments)-1 > 360 {
		return fmt.Errorf("%w: the updated schedule must fit within 360 installments", ErrInvalidSchedule)
	}
	if err := validateInstallments(name, sourceID, uuid.Nil, installments); err != nil {
		return err
	}
	return s.scheduleRepo.UpdateInstallments(ctx, userID, scheduleID, sourceID, name, firstNumber, installments)
}

func (s *Service) DeleteInstallments(ctx context.Context, userID, scheduleID uuid.UUID) error {
	return s.scheduleRepo.DeleteInstallments(ctx, userID, scheduleID)
}

func (s *Service) PayCardInstallment(ctx context.Context, userID, occurrenceID uuid.UUID, at time.Time) (uuid.UUID, error) {
	txID, cardID, sourceID, err := s.scheduleRepo.PayCardInstallment(ctx, userID, occurrenceID, at)
	if err == nil {
		cache.InvalidateCreditCard(ctx, s.redis, cardID, userID, sourceID)
	}
	return txID, err
}

func validateInstallments(name string, sourceID, accountID uuid.UUID, installments []Installment) error {
	switch {
	case name == "":
		return fmt.Errorf("%w: name is required", ErrInvalidSchedule)
	case sourceID == accountID:
		return fmt.Errorf("%w: pay from a different account than the loan", ErrInvalidSchedule)
	case len(installments) == 0:
		return fmt.Errorf("%w: add at least one installment", ErrInvalidSchedule)
	case len(installments) > 360:
		return fmt.Errorf("%w: a schedule can have at most 360 installments", ErrInvalidSchedule)
	}
	for i, in := range installments {
		if !in.Amount.IsPositive() {
			return fmt.Errorf("%w: installment %d must be more than zero", ErrInvalidSchedule, i+1)
		}
		if i > 0 && !in.DueDate.After(installments[i-1].DueDate) {
			return fmt.Errorf("%w: installment %d must be due after installment %d", ErrInvalidSchedule, i+1, i)
		}
	}
	return nil
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
