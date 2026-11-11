package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kaloy/finankal/engine-go/finance"
	"github.com/kaloy/finankal/engine-go/internal/credit"
	"github.com/kaloy/finankal/engine-go/internal/db"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/kaloy/finankal/engine-go/internal/loan"
	"github.com/kaloy/finankal/engine-go/internal/schedule"
	"github.com/kaloy/finankal/engine-go/internal/statement"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const statementDateFormat = "2006-01-02"

type Server struct {
	finance.UnimplementedFinanceEngineServer

	ledgerService    *ledger.Service
	creditService    *credit.Service
	scheduleService  *schedule.Service
	statementService *statement.Service
	loanService      *loan.Service
}

func NewServer(ledgerService *ledger.Service, creditService *credit.Service,
	scheduleService *schedule.Service, statementService *statement.Service, loanService *loan.Service) *Server {

	return &Server{
		ledgerService:    ledgerService,
		creditService:    creditService,
		scheduleService:  scheduleService,
		statementService: statementService,
		loanService:      loanService,
	}
}

// toStatus maps a service error to a gRPC status. Unexpected errors are logged
// here and not returned, because they can contain SQL and table names.
func toStatus(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return status.Error(codes.NotFound, "not found")

	case errors.Is(err, ledger.ErrInvalidTransaction),
		errors.Is(err, credit.ErrInvalidRequest),
		errors.Is(err, schedule.ErrInvalidSchedule),
		errors.Is(err, statement.ErrInvalidStatement),
		errors.Is(err, loan.ErrInvalidLoan):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, statement.ErrAlreadyImported),
		errors.Is(err, loan.ErrAlreadyImported):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, credit.ErrCreditLimitExceeded),
		errors.Is(err, credit.ErrStatementAlreadyPaid),
		errors.Is(err, credit.ErrPaymentTooLarge):
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	db.LogPgError(err)

	return status.Error(codes.Internal, "internal error")
}

func parseID(value string, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, status.Errorf(codes.InvalidArgument, "invalid %s", field)
	}

	return id, nil
}

func parseAmount(value string, field string) (decimal.Decimal, error) {
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, status.Errorf(codes.InvalidArgument, "invalid %s", field)
	}

	return amount, nil
}

// parseOptionalAmount treats "" as "not printed on the statement".
func parseOptionalAmount(value string, field string) (decimal.NullDecimal, error) {
	if value == "" {
		return decimal.NullDecimal{}, nil
	}
	amount, err := parseAmount(value, field)
	return decimal.NullDecimal{Decimal: amount, Valid: err == nil}, err
}

// parseDate reads a yyyy-MM-dd calendar date.
func parseDate(value string, field string) (time.Time, error) {
	date, err := time.Parse(statementDateFormat, value)
	if err != nil {
		return time.Time{}, status.Errorf(codes.InvalidArgument, "invalid %s", field)
	}

	return date, nil
}

func clampLimit(limit int32, fallback int, max int) int {
	if limit <= 0 {
		return fallback
	}

	return min(int(limit), max)
}

func toProtoTransaction(transaction ledger.TransactionDetail) *finance.TransactionResponse {
	response := &finance.TransactionResponse{
		Id:          transaction.ID.String(),
		Description: transaction.Description,
		CreatedAt:   timestamppb.New(transaction.CreatedAt),
	}

	for _, entry := range transaction.Entries {
		response.Entries = append(response.Entries, &finance.TransactionEntry{
			Id:            entry.ID.String(),
			TransactionId: entry.TransactionID.String(),
			AccountId:     entry.AccountID.String(),
			AccountName:   entry.AccountName,
			AccountType:   string(entry.AccountType),
			Amount:        entry.Amount.String(),
			Type:          string(entry.Type),
		})
	}

	return response
}

// ===============================
// HEALTH
// ===============================

func (s *Server) HealthCheck(ctx context.Context, req *finance.HealthRequest) (*finance.HealthResponse, error) {
	return &finance.HealthResponse{Status: "OK"}, nil
}

// ===============================
// LEDGER
// ===============================

func (s *Server) CreateTransaction(ctx context.Context, req *finance.CreateTransactionRequest) (*finance.CreateTransactionResponse, error) {
	entries := make([]ledger.Entry, len(req.Entries))

	for i, e := range req.Entries {
		accountID, err := parseID(e.AccountId, "account id")
		if err != nil {
			return nil, err
		}

		amount, err := parseAmount(e.Amount, "amount")
		if err != nil {
			return nil, err
		}

		entries[i] = ledger.Entry{
			AccountID: accountID,
			Amount:    amount,
			Type:      ledger.EntryType(e.Type),
		}
	}

	var occurredAt *time.Time
	if req.OccurredAt != nil {
		if err := req.OccurredAt.CheckValid(); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid occurred_at: %v", err)
		}
		t := req.OccurredAt.AsTime()
		occurredAt = &t
	}

	txID, err := s.ledgerService.CreateTransaction(ctx, req.Description, entries, occurredAt)
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.CreateTransactionResponse{TransactionId: txID.String()}, nil
}

func (s *Server) GetBalance(ctx context.Context, req *finance.GetBalanceRequest) (*finance.GetBalanceResponse, error) {
	accountID, err := parseID(req.AccountId, "account id")
	if err != nil {
		return nil, err
	}

	balance, err := s.ledgerService.GetBalance(ctx, accountID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.GetBalanceResponse{Balance: balance.String()}, nil
}

func (s *Server) GetAccountSummary(ctx context.Context, req *finance.GetAccountSummaryRequest) (*finance.GetAccountSummaryResponse, error) {
	accountID, err := parseID(req.AccountId, "account id")
	if err != nil {
		return nil, err
	}

	summary, err := s.ledgerService.GetAccountSummary(ctx, accountID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.GetAccountSummaryResponse{
		Account: &finance.Account{
			Id:        summary.AccountID.String(),
			Name:      summary.Name,
			Type:      string(summary.Type),
			Balance:   summary.Balance.String(),
			CreatedAt: timestamppb.New(summary.CreatedAt),
		},
	}, nil
}

func (s *Server) ImportBankStatement(ctx context.Context, req *finance.ImportBankStatementRequest) (*finance.ImportBankStatementResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	var accountID uuid.UUID // optional: found by account number when empty
	if req.GetAccountId() != "" {
		if accountID, err = parseID(req.GetAccountId(), "account id"); err != nil {
			return nil, err
		}
	}
	periodStart, err := parseDate(req.GetPeriodStart(), "period start")
	if err != nil {
		return nil, err
	}
	periodEnd, err := parseDate(req.GetPeriodEnd(), "period end")
	if err != nil {
		return nil, err
	}
	opening, err := parseOptionalAmount(req.GetOpeningBalance(), "opening balance")
	if err != nil {
		return nil, err
	}
	closing, err := parseOptionalAmount(req.GetClosingBalance(), "closing balance")
	if err != nil {
		return nil, err
	}

	rows := make([]statement.Row, len(req.GetRows()))
	for i, r := range req.GetRows() {
		date, err := parseDate(r.GetDate(), "row date")
		if err != nil {
			return nil, err
		}
		amount, err := parseAmount(r.GetAmount(), "row amount")
		if err != nil {
			return nil, err
		}
		rows[i] = statement.Row{Date: date, Description: r.GetDescription(), Amount: amount, MoneyOut: r.GetMoneyOut(), Category: r.GetCategory()}
	}

	// A credit card statement's bill, when the statement printed one.
	var bill *statement.Bill
	if req.GetStatementDate() != "" {
		statementDate, err := parseDate(req.GetStatementDate(), "statement date")
		if err != nil {
			return nil, err
		}
		totalDue, err := parseAmount(req.GetTotalDue(), "total due")
		if err != nil {
			return nil, err
		}
		bill = &statement.Bill{StatementDate: statementDate, TotalDue: totalDue}
		if req.GetDueDate() != "" {
			dueDate, err := parseDate(req.GetDueDate(), "due date")
			if err != nil {
				return nil, err
			}
			bill.DueDate = &dueDate
		}
	}

	result, err := s.statementService.Import(ctx, statement.Statement{
		Bill:           bill,
		UserID:         userID,
		AccountID:      accountID,
		AccountNumber:  req.GetAccountNumber(),
		AccountName:    req.GetAccountName(),
		AccountType:    req.GetAccountType(),
		Bank:           req.GetBank(),
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		OpeningBalance: opening,
		ClosingBalance: closing,
		FileSHA256:     req.GetFileSha256(),
		Rows:           rows,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.ImportBankStatementResponse{
		StatementId: result.StatementID.String(),
		Imported:    int32(result.Imported),
		Warnings:    result.Warnings,
		Duplicates:  result.Duplicates,
		AccountId:   result.AccountID.String(),
		AccountName: result.AccountName,
		Transfers:   result.Transfers,
	}, nil
}

func (s *Server) ListBankStatements(ctx context.Context, req *finance.ListBankStatementsRequest) (*finance.ListBankStatementsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}

	statements, err := s.statementService.List(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	resp := &finance.ListBankStatementsResponse{}
	for _, st := range statements {
		resp.Statements = append(resp.Statements, &finance.BankStatementSummary{
			Id:          st.ID.String(),
			AccountId:   st.AccountID.String(),
			Bank:        st.Bank,
			PeriodStart: st.PeriodStart.Format(statementDateFormat),
			PeriodEnd:   st.PeriodEnd.Format(statementDateFormat),
			ImportedAt:  timestamppb.New(st.ImportedAt),
		})
	}
	return resp, nil
}

func (s *Server) ClearUserTransactions(ctx context.Context, req *finance.ClearUserTransactionsRequest) (*finance.ClearUserTransactionsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}

	deleted, err := s.ledgerService.ClearTransactions(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.ClearUserTransactionsResponse{DeletedTransactions: deleted}, nil
}

func (s *Server) GetPayables(ctx context.Context, req *finance.GetPayablesRequest) (*finance.GetPayablesResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}

	// Dates are calendar days (Philippine calendar), like the bills' and installments'.
	now := time.Now().In(time.FixedZone("PHT", 8*60*60))
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Loans charged on what's owed get their interest booked up to today before anything is read.
	if err := s.loanService.Accrue(ctx, userID, today); err != nil {
		return nil, toStatus(err)
	}
	payables, err := s.scheduleService.GetPayables(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}
	if payables, err = s.loanService.Settle(ctx, userID, payables); err != nil {
		return nil, toStatus(err)
	}

	deref := func(n *int) int32 {
		if n == nil {
			return 0
		}
		return int32(*n)
	}
	links, err := s.creditService.PaikotLinks(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}
	// Paikot: where the due lands if the card linked to its account pays it.
	roll := func(p *finance.Payable, accountID uuid.UUID, due time.Time, amount decimal.Decimal) {
		link, ok := links[accountID]
		if !ok {
			return
		}
		r := link.Roll(due, today, amount)
		p.PaikotCardId = link.Card.CardID.String()
		p.PaikotCardName = link.CardName
		p.PaikotPayFrom = r.PayFrom.Format(statementDateFormat)
		p.PaikotPayBy = r.PayBy.Format(statementDateFormat)
		p.PaikotDueDate = r.DueDate.Format(statementDateFormat)
		p.PaikotFee = r.Fee.String()
	}
	// Card bills still owed, plus each card's current cycle so far, alongside the loan installments.
	bills, err := s.creditService.OpenBills(ctx, userID, today)
	if err != nil {
		return nil, toStatus(err)
	}
	current := map[uuid.UUID]*credit.Bill{} // each card's cycle so far
	for i := range bills {
		if bills[i].SoFar {
			current[bills[i].CardID] = &bills[i]
		}
	}
	withInstallments := map[uuid.UUID]bool{}

	resp := &finance.GetPayablesResponse{}
	for _, p := range payables {
		// A card installment is billed on the statement its date falls in: one in the current cycle is part
		// of that bill, an earlier one is already on a statement.
		// shortcut: from a statement date until that statement is imported, its installments show in neither; fold by imported statements if the gap matters.
		if b := current[p.AccountID]; b != nil && !p.DueDate.After(b.StatementDate) {
			if !p.DueDate.Before(b.StartDate) {
				b.Owed = b.Owed.Add(p.Amount)
				withInstallments[p.AccountID] = true
			}
			continue
		}
		dueDate := p.DueDate
		if p.AccountType == string(ledger.CREDIT_CARD) {
			card, err := s.creditService.GetCreditCard(ctx, p.AccountID)
			if err != nil {
				return nil, toStatus(err)
			}
			dueDate = credit.CalculateBillingCycle(p.DueDate, *card).DueDate
		}
		payable := &finance.Payable{
			Name:              p.Name,
			ScheduleId:        p.ScheduleID.String(),
			OccurrenceId:      p.OccurrenceID.String(),
			AccountId:         p.AccountID.String(),
			AccountName:       p.AccountName,
			DueDate:           dueDate.Format(statementDateFormat),
			Amount:            p.Amount.String(),
			InstallmentNumber: deref(p.InstallmentNumber),
			Installments:      deref(p.Installments),
			ScheduledDate:     p.DueDate.Format(statementDateFormat),
		}
		if p.SourceAccountID != nil {
			payable.SourceAccountId = p.SourceAccountID.String()
		}
		roll(payable, p.AccountID, dueDate, p.Amount)
		resp.Payables = append(resp.Payables, payable)
	}

	for _, b := range bills {
		soFar := ""
		if b.SoFar {
			soFar = ", so far"
			if withInstallments[b.CardID] {
				soFar = ", so far, with installments"
			}
		}
		payable := &finance.Payable{
			Name:        fmt.Sprintf("%s bill (%s – %s%s)", b.CardName, b.StartDate.Format("2 Jan"), b.StatementDate.Format("2 Jan"), soFar),
			AccountId:   b.CardID.String(),
			AccountName: b.CardName,
			DueDate:     b.DueDate.Format(statementDateFormat),
			Amount:      b.Owed.String(),
		}
		roll(payable, b.CardID, b.DueDate, b.Owed)
		resp.Payables = append(resp.Payables, payable)
	}
	sort.SliceStable(resp.Payables, func(i, j int) bool { return resp.Payables[i].DueDate < resp.Payables[j].DueDate })
	return resp, nil
}

func (s *Server) SetPaikotLink(ctx context.Context, req *finance.SetPaikotLinkRequest) (*finance.SetPaikotLinkResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	accountID, err := parseID(req.GetAccountId(), "account id")
	if err != nil {
		return nil, err
	}
	var cardID uuid.UUID // empty removes the link
	fee := decimal.Zero
	if req.GetCardId() != "" {
		if cardID, err = parseID(req.GetCardId(), "card id"); err != nil {
			return nil, err
		}
		if req.GetFeePercent() != "" {
			if fee, err = parseAmount(req.GetFeePercent(), "fee percent"); err != nil {
				return nil, err
			}
		}
	}
	if err := s.creditService.SetPaikotLink(ctx, userID, accountID, cardID, fee); err != nil {
		return nil, toStatus(err)
	}
	return &finance.SetPaikotLinkResponse{}, nil
}

func (s *Server) ImportLoan(ctx context.Context, req *finance.ImportLoanRequest) (*finance.ImportLoanResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	var sourceID uuid.UUID // optional
	if req.GetSourceAccountId() != "" {
		if sourceID, err = parseID(req.GetSourceAccountId(), "source account id"); err != nil {
			return nil, err
		}
	}
	bookedOn, err := parseDate(req.GetBookedOn(), "booking date")
	if err != nil {
		return nil, err
	}
	amount, err := parseAmount(req.GetAmount(), "loan amount")
	if err != nil {
		return nil, err
	}
	deducted, err := parseAmount(req.GetDeducted(), "deductions")
	if err != nil {
		return nil, err
	}
	rate, err := parseAmount(req.GetDailyRate(), "daily rate")
	if err != nil {
		return nil, err
	}
	installments := make([]schedule.Installment, 0, len(req.GetInstallments()))
	for _, in := range req.GetInstallments() {
		due, err := parseDate(in.GetDueDate(), "installment due date")
		if err != nil {
			return nil, err
		}
		amt, err := parseAmount(in.GetAmount(), "installment amount")
		if err != nil {
			return nil, err
		}
		installments = append(installments, schedule.Installment{DueDate: due, Amount: amt})
	}

	imported, err := s.loanService.Import(ctx, loan.Import{
		UserID:        userID,
		AccountNumber: req.GetAccountNumber(),
		AccountName:   req.GetAccountName(),
		BookedOn:      bookedOn,
		Amount:        amount,
		Deducted:      deducted,
		DailyRate:     rate,
		SourceID:      sourceID,
		ScheduleName:  req.GetScheduleName(),
		Installments:  installments,
		FileSHA256:    req.GetFileSha256(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &finance.ImportLoanResponse{AccountId: imported.AccountID.String(), AccountName: imported.AccountName}, nil
}

func (s *Server) CreateInstallments(ctx context.Context, req *finance.CreateInstallmentsRequest) (*finance.CreateInstallmentsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceAccountId(), "source account id")
	if err != nil {
		return nil, err
	}
	accountID, err := parseID(req.GetAccountId(), "account id")
	if err != nil {
		return nil, err
	}

	installments := make([]schedule.Installment, 0, len(req.GetInstallments()))
	for _, in := range req.GetInstallments() {
		due, err := time.Parse(statementDateFormat, in.GetDueDate())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid due date %q", in.GetDueDate())
		}
		amount, err := parseAmount(in.GetAmount(), "installment amount")
		if err != nil {
			return nil, err
		}
		installments = append(installments, schedule.Installment{DueDate: due, Amount: amount})
	}

	id, err := s.scheduleService.CreateInstallments(ctx, userID, req.GetName(), sourceID, accountID, installments)
	if err != nil {
		return nil, toStatus(err)
	}
	return &finance.CreateInstallmentsResponse{ScheduleId: id.String()}, nil
}

func (s *Server) UpdateInstallments(ctx context.Context, req *finance.UpdateInstallmentsRequest) (*finance.UpdateInstallmentsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	scheduleID, err := parseID(req.GetScheduleId(), "schedule id")
	if err != nil {
		return nil, err
	}
	firstNumber := int(req.GetFirstInstallmentNumber())
	now := time.Now().In(time.FixedZone("PHT", 8*60*60))
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if err := s.loanService.Accrue(ctx, userID, today); err != nil {
		return nil, toStatus(err)
	}
	stored, err := s.scheduleService.GetPayables(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}
	payables, err := s.loanService.Settle(ctx, userID, stored)
	if err != nil {
		return nil, toStatus(err)
	}
	firstPending := 0
	// A partly paid first installment is shown at what's left of it; left and full are those two amounts.
	var left, full decimal.Decimal
	for _, p := range payables {
		if p.ScheduleID == scheduleID && p.InstallmentNumber != nil {
			firstPending = *p.InstallmentNumber
			for _, st := range stored {
				if st.OccurrenceID == p.OccurrenceID {
					left, full = p.Amount, st.Amount
				}
			}
			break
		}
	}
	// A card plan's first pending installment can be on the current bill, so the editor starts after it.
	if firstPending == 0 || firstNumber < firstPending {
		return nil, status.Error(codes.FailedPrecondition, "refresh the plan before editing its remaining installments")
	}
	sourceID, err := parseID(req.GetSourceAccountId(), "source account id")
	if err != nil {
		return nil, err
	}
	items := make([]schedule.Installment, 0, len(req.GetInstallments()))
	for _, in := range req.GetInstallments() {
		due, e := parseDate(in.GetDueDate(), "installment due date")
		if e != nil {
			return nil, e
		}
		amount, e := parseAmount(in.GetAmount(), "installment amount")
		if e != nil {
			return nil, e
		}
		items = append(items, schedule.Installment{DueDate: due, Amount: amount})
	}
	// Payments are taken off the stored amount on every read, so saving what's left would count them twice.
	if len(items) > 0 && firstNumber == firstPending && !left.Equal(full) {
		if !items[0].Amount.Equal(left) {
			return nil, status.Error(codes.FailedPrecondition, "the first installment is partly paid; leave its amount as shown")
		}
		items[0].Amount = full
	}
	if err := s.scheduleService.UpdateInstallments(ctx, userID, scheduleID, sourceID, req.GetName(), firstNumber, items); err != nil {
		return nil, toStatus(err)
	}
	return &finance.UpdateInstallmentsResponse{}, nil
}

func (s *Server) DeleteInstallments(ctx context.Context, req *finance.DeleteInstallmentsRequest) (*finance.DeleteInstallmentsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	id, err := parseID(req.GetScheduleId(), "schedule id")
	if err != nil {
		return nil, err
	}
	if err := s.scheduleService.DeleteInstallments(ctx, userID, id); err != nil {
		return nil, toStatus(err)
	}
	return &finance.DeleteInstallmentsResponse{}, nil
}

func (s *Server) PayInstallment(ctx context.Context, req *finance.PayInstallmentRequest) (*finance.PayInstallmentResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	occurrenceID, err := parseID(req.GetOccurrenceId(), "occurrence id")
	if err != nil {
		return nil, err
	}
	now := time.Now().In(time.FixedZone("PHT", 8*60*60))
	id, err := s.scheduleService.PayCardInstallment(ctx, userID, occurrenceID, now)
	if err != nil {
		return nil, toStatus(err)
	}
	return &finance.PayInstallmentResponse{TransactionId: id.String()}, nil
}

func (s *Server) GetLoan(ctx context.Context, req *finance.GetLoanRequest) (*finance.LoanDetails, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	accountID, err := parseID(req.GetAccountId(), "account id")
	if err != nil {
		return nil, err
	}
	d, err := s.loanService.Get(ctx, userID, accountID)
	if err != nil {
		return nil, toStatus(err)
	}
	return &finance.LoanDetails{AccountId: d.AccountID.String(), Name: d.Name, BookedOn: d.BookedOn.Format(statementDateFormat), DailyRate: d.DailyRate.String(), PrincipalOn: d.PrincipalOn.Format(statementDateFormat), Principal: d.Principal.String()}, nil
}

func (s *Server) UpdateLoan(ctx context.Context, req *finance.UpdateLoanRequest) (*finance.UpdateLoanResponse, error) {
	d := req.GetLoan()
	if d == nil {
		return nil, status.Error(codes.InvalidArgument, "loan details are required")
	}
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	accountID, err := parseID(d.GetAccountId(), "account id")
	if err != nil {
		return nil, err
	}
	booked, err := parseDate(d.GetBookedOn(), "booking date")
	if err != nil {
		return nil, err
	}
	principalOn, err := parseDate(d.GetPrincipalOn(), "principal date")
	if err != nil {
		return nil, err
	}
	rate, err := parseAmount(d.GetDailyRate(), "daily rate")
	if err != nil {
		return nil, err
	}
	principal, err := parseAmount(d.GetPrincipal(), "principal")
	if err != nil {
		return nil, err
	}
	if err := s.loanService.Update(ctx, userID, loan.Details{AccountID: accountID, Name: d.GetName(), BookedOn: booked, DailyRate: rate, PrincipalOn: principalOn, Principal: principal}); err != nil {
		return nil, toStatus(err)
	}
	return &finance.UpdateLoanResponse{}, nil
}

func (s *Server) DeleteLoan(ctx context.Context, req *finance.DeleteLoanRequest) (*finance.DeleteLoanResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	accountID, err := parseID(req.GetAccountId(), "account id")
	if err != nil {
		return nil, err
	}
	if err := s.loanService.Delete(ctx, userID, accountID); err != nil {
		return nil, toStatus(err)
	}
	return &finance.DeleteLoanResponse{}, nil
}

func (s *Server) GetMoneyFlows(ctx context.Context, req *finance.GetMoneyFlowsRequest) (*finance.GetMoneyFlowsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	// Philippine calendar days; `to` is inclusive, so the query's upper bound is the next midnight.
	manila := time.FixedZone("PHT", 8*60*60)
	var from, to *time.Time
	if req.GetFrom() != "" {
		day, err := parseDate(req.GetFrom(), "from")
		if err != nil {
			return nil, err
		}
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, manila)
		from = &start
	}
	if req.GetTo() != "" {
		day, err := parseDate(req.GetTo(), "to")
		if err != nil {
			return nil, err
		}
		end := time.Date(day.Year(), day.Month(), day.Day()+1, 0, 0, 0, 0, manila)
		to = &end
	}

	flows, err := s.ledgerService.GetMoneyFlows(ctx, userID, from, to)
	if err != nil {
		return nil, toStatus(err)
	}

	resp := &finance.GetMoneyFlowsResponse{}
	for _, f := range flows {
		resp.Flows = append(resp.Flows, &finance.MoneyFlow{
			FromAccountId: f.FromAccountID.String(),
			ToAccountId:   f.ToAccountID.String(),
			Amount:        f.Amount.String(),
			Transactions:  int32(f.Transactions),
		})
	}
	return resp, nil
}

func (s *Server) GetUserAccounts(ctx context.Context, req *finance.GetUserAccountsRequest) (*finance.GetUserAccountsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}

	accounts, err := s.ledgerService.GetUserAccounts(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}
	cards, err := s.creditService.ListCreditCards(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}
	cardTerms := make(map[uuid.UUID]credit.CreditCard, len(cards))
	for _, card := range cards {
		cardTerms[card.CardID] = card
	}

	resp := &finance.GetUserAccountsResponse{
		Accounts: make([]*finance.Account, 0, len(accounts)),
	}

	for _, account := range accounts {
		accountDto := &finance.Account{
			Id:        account.ID.String(),
			Name:      account.Name,
			Type:      account.Type,
			Balance:   account.Balance.String(),
			CreatedAt: timestamppb.New(account.CreatedAt),
		}
		if card, ok := cardTerms[account.ID]; ok {
			accountDto.CreditLimit = card.CreditLimit.String()
			accountDto.BillingDay = int32(card.BillingDay)
			accountDto.StatementShift = card.StatementShift
		}
		resp.Accounts = append(resp.Accounts, accountDto)
	}

	return resp, nil
}

func (s *Server) GetLedgerEntries(ctx context.Context, req *finance.GetLedgerEntriesRequest) (*finance.GetLedgerEntriesResponse, error) {
	accountID, err := parseID(req.AccountId, "account id")
	if err != nil {
		return nil, err
	}

	entries, err := s.ledgerService.GetLedgerEntries(ctx, accountID)
	if err != nil {
		return nil, toStatus(err)
	}

	response := &finance.GetLedgerEntriesResponse{}

	for _, e := range entries {
		response.Entries = append(response.Entries, &finance.LedgerEntry{
			TransactionId: e.TransactionID.String(),
			AccountId:     e.AccountID.String(),
			Amount:        e.Amount.String(),
			Type:          string(e.Type),
			Description:   e.Description,
			CreatedAt:     e.CreatedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func (s *Server) GetUserTotalCredit(ctx context.Context, req *finance.GetUserTotalRequest) (*finance.GetUserTotalResponse, error) {
	userID, err := parseID(req.UserId, "user id")
	if err != nil {
		return nil, err
	}

	total, err := s.ledgerService.GetUserTotalCredit(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.GetUserTotalResponse{UserId: req.UserId, TotalCredit: total.String()}, nil
}

func (s *Server) GetUserTotalDebit(ctx context.Context, req *finance.GetUserTotalRequest) (*finance.GetUserTotalResponse, error) {
	userID, err := parseID(req.UserId, "user id")
	if err != nil {
		return nil, err
	}

	total, err := s.ledgerService.GetUserTotalDebit(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.GetUserTotalResponse{UserId: req.UserId, TotalDebit: total.String()}, nil
}

func (s *Server) GetUserNetWorth(ctx context.Context, req *finance.GetUserTotalRequest) (*finance.GetUserTotalResponse, error) {
	userID, err := parseID(req.UserId, "user id")
	if err != nil {
		return nil, err
	}

	total, err := s.ledgerService.GetUserNetWorth(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.GetUserTotalResponse{UserId: req.UserId, TotalBalance: total.String()}, nil
}

func (s *Server) GetRecentTransactions(ctx context.Context, req *finance.GetRecentTransactionsRequest) (*finance.GetRecentTransactionsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}

	transactions, err := s.ledgerService.GetUserRecentTransactions(ctx, userID, clampLimit(req.GetLimit(), 5, 20))
	if err != nil {
		return nil, toStatus(err)
	}

	response := &finance.GetRecentTransactionsResponse{}

	for _, transaction := range transactions {
		response.Transactions = append(response.Transactions, &finance.RecentTransaction{
			Id:          transaction.ID.String(),
			Description: transaction.Description,
			Amount:      transaction.Amount.String(),
			Direction:   transaction.Direction,
			AccountName: transaction.AccountName,
			CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}

func (s *Server) GetTransaction(ctx context.Context, req *finance.GetTransactionRequest) (*finance.TransactionResponse, error) {
	transactionID, err := parseID(req.GetTransactionId(), "transaction id")
	if err != nil {
		return nil, err
	}

	transaction, err := s.ledgerService.GetTransaction(ctx, transactionID)
	if err != nil {
		return nil, toStatus(err)
	}

	return toProtoTransaction(*transaction), nil
}

func (s *Server) UpdateTransaction(ctx context.Context, req *finance.UpdateTransactionRequest) (*finance.UpdateTransactionResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	transactionID, err := parseID(req.GetTransactionId(), "transaction id")
	if err != nil {
		return nil, err
	}

	entries := make([]ledger.Entry, len(req.Entries))
	for i, e := range req.Entries {
		entryID, err := parseID(e.Id, "entry id")
		if err != nil {
			return nil, err
		}
		accountID, err := parseID(e.AccountId, "account id")
		if err != nil {
			return nil, err
		}
		amount, err := parseAmount(e.Amount, "amount")
		if err != nil {
			return nil, err
		}
		entries[i] = ledger.Entry{ID: entryID, AccountID: accountID, Amount: amount, Type: ledger.EntryType(e.Type)}
	}

	var occurredAt *time.Time
	if req.OccurredAt != nil {
		if err := req.OccurredAt.CheckValid(); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid occurred_at: %v", err)
		}
		t := req.OccurredAt.AsTime()
		occurredAt = &t
	}

	if err := s.ledgerService.UpdateTransaction(ctx, userID, transactionID, req.Description, occurredAt, entries); err != nil {
		return nil, toStatus(err)
	}
	return &finance.UpdateTransactionResponse{}, nil
}

func (s *Server) DeleteTransaction(ctx context.Context, req *finance.DeleteTransactionRequest) (*finance.DeleteTransactionResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}
	transactionID, err := parseID(req.GetTransactionId(), "transaction id")
	if err != nil {
		return nil, err
	}

	if err := s.ledgerService.DeleteTransaction(ctx, userID, transactionID); err != nil {
		return nil, toStatus(err)
	}
	return &finance.DeleteTransactionResponse{}, nil
}

func (s *Server) GetUserTransactions(ctx context.Context, req *finance.GetUserTransactionsRequest) (*finance.GetUserTransactionsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}

	offset := max(int(req.GetOffset()), 0)

	transactions, total, err := s.ledgerService.GetUserTransactions(ctx, userID, clampLimit(req.GetLimit(), 50, 100), offset)
	if err != nil {
		return nil, toStatus(err)
	}

	response := &finance.GetUserTransactionsResponse{Total: int32(total)}

	for _, transaction := range transactions {
		response.Transactions = append(response.Transactions, toProtoTransaction(transaction))
	}

	return response, nil
}

// ===============================
// CREDIT CARD
// ===============================

func (s *Server) CreateCreditCard(ctx context.Context, req *finance.CreateCreditCardRequest) (*finance.CreateCreditCardResponse, error) {
	var accountID uuid.UUID
	var err error

	// Optional: an empty id asks for a new account.
	if req.AccountId != "" {
		accountID, err = parseID(req.AccountId, "account id")
		if err != nil {
			return nil, err
		}
	}

	limit, err := parseAmount(req.CreditLimit, "credit limit")
	if err != nil {
		return nil, err
	}

	cardID, err := s.creditService.CreateCreditCard(ctx, accountID, req.AccountName, limit, int(req.BillingDay), int(req.PaymentDueDays))
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.CreateCreditCardResponse{CreditCardId: cardID.String()}, nil
}

func (s *Server) RecordCreditCardTransaction(ctx context.Context, req *finance.RecordCreditCardTransactionRequest) (*finance.RecordCreditCardTransactionResponse, error) {
	cardID, err := parseID(req.CardId, "card id")
	if err != nil {
		return nil, err
	}

	amount, err := parseAmount(req.Amount, "amount")
	if err != nil {
		return nil, err
	}

	// A nil timestamp converts to 1970 and would bill the purchase to a 1970 statement.
	if req.PurchaseDate == nil {
		return nil, status.Error(codes.InvalidArgument, "purchase date is required")
	}

	txID, err := s.creditService.RecordCreditCardTransaction(ctx, credit.CreditCardTransactionRequest{
		CardID:       cardID,
		Amount:       amount,
		Description:  req.Description,
		PurchaseDate: req.PurchaseDate.AsTime(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.RecordCreditCardTransactionResponse{TransactionId: txID.String()}, nil
}

func (s *Server) PayCreditCardStatement(ctx context.Context, req *finance.PayCreditCardStatementRequest) (*finance.PayCreditCardStatementResponse, error) {
	statementID, err := parseID(req.StatementId, "statement id")
	if err != nil {
		return nil, err
	}

	cardID, err := parseID(req.CardId, "card id")
	if err != nil {
		return nil, err
	}

	paymentAccountID, err := parseID(req.PaymentAccountId, "payment account id")
	if err != nil {
		return nil, err
	}

	amount, err := parseAmount(req.Amount, "amount")
	if err != nil {
		return nil, err
	}

	txID, err := s.creditService.PayCreditCardStatement(ctx, credit.CreditCardPaymentRequest{
		StatementID:      statementID,
		CardID:           cardID,
		PaymentAccountID: paymentAccountID,
		Amount:           amount,
		Description:      req.Description,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.PayCreditCardStatementResponse{TransactionId: txID.String()}, nil
}

func (s *Server) GetCreditCardStatement(ctx context.Context, req *finance.GetCreditCardStatementRequest) (*finance.GetCreditCardStatementResponse, error) {
	statementID, err := parseID(req.StatementId, "statement id")
	if err != nil {
		return nil, err
	}

	statement, err := s.creditService.GetStatement(ctx, statementID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &finance.GetCreditCardStatementResponse{
		StatementId:   statement.ID.String(),
		CreditCardId:  statement.CardID.String(),
		StartDate:     statement.StartDate.Format(statementDateFormat),
		EndDate:       statement.EndDate.Format(statementDateFormat),
		StatementDate: statement.StatementDate.Format(statementDateFormat),
		DueDate:       statement.DueDate.Format(statementDateFormat),
		TotalAmount:   statement.TotalAmount.String(),
		Status:        string(statement.Status),
	}, nil
}
