package main

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kaloy/finankal/engine-go/finance"
	"github.com/kaloy/finankal/engine-go/internal/credit"
	"github.com/kaloy/finankal/engine-go/internal/db"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/kaloy/finankal/engine-go/internal/schedule"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const statementDateFormat = "2006-01-02"

type Server struct {
	finance.UnimplementedFinanceEngineServer

	ledgerService   *ledger.Service
	creditService   *credit.Service
	scheduleService *schedule.Service
}

func NewServer(ledgerService *ledger.Service, creditService *credit.Service,
	scheduleService *schedule.Service) *Server {

	return &Server{
		ledgerService:   ledgerService,
		creditService:   creditService,
		scheduleService: scheduleService,
	}
}

// toStatus maps a service error to a gRPC status. Unexpected errors are logged
// here and not returned, because they can contain SQL and table names.
func toStatus(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return status.Error(codes.NotFound, "not found")

	case errors.Is(err, ledger.ErrInvalidTransaction),
		errors.Is(err, credit.ErrInvalidRequest):
		return status.Error(codes.InvalidArgument, err.Error())

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

	txID, err := s.ledgerService.CreateTransaction(ctx, req.Description, entries)
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

func (s *Server) GetUserAccounts(ctx context.Context, req *finance.GetUserAccountsRequest) (*finance.GetUserAccountsResponse, error) {
	userID, err := parseID(req.GetUserId(), "user id")
	if err != nil {
		return nil, err
	}

	accounts, err := s.ledgerService.GetUserAccounts(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	resp := &finance.GetUserAccountsResponse{
		Accounts: make([]*finance.Account, 0, len(accounts)),
	}

	for _, account := range accounts {
		resp.Accounts = append(resp.Accounts, &finance.Account{
			Id:        account.ID.String(),
			Name:      account.Name,
			Type:      account.Type,
			Balance:   account.Balance.String(),
			CreatedAt: timestamppb.New(account.CreatedAt),
		})
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
