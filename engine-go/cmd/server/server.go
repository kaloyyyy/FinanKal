package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/finance"
	"github.com/kaloy/finankal/engine-go/internal/credit"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/shopspring/decimal"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	finance.UnimplementedFinanceEngineServer

	ledgerService *ledger.Service
	creditService *credit.Service
}

func NewServer(ledgerService *ledger.Service, creditService *credit.Service) *server {

	return &server{
		ledgerService: ledgerService,
		creditService: creditService,
	}
}

// ===============================
// HEALTH
// ===============================

func (s *server) HealthCheck(ctx context.Context, req *finance.HealthRequest) (*finance.HealthResponse, error) {

	return &finance.HealthResponse{
		Status: "OK",
	}, nil
}

// ===============================
// LEDGER
// ===============================

func (s *server) CreateTransaction(ctx context.Context, req *finance.CreateTransactionRequest) (*finance.CreateTransactionResponse, error) {

	entries := make([]ledger.Entry, len(req.Entries))

	for i, e := range req.Entries {

		accountID, err := uuid.Parse(e.AccountId)

		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid account id")
		}

		amount, err := decimal.NewFromString(e.Amount)

		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid amount")
		}

		entryType := ledger.EntryType(e.Type)

		entries[i] = ledger.Entry{
			AccountID: accountID,
			Amount:    amount,
			Type:      entryType,
		}
	}

	txID, err := s.ledgerService.CreateTransaction(ctx, req.Description, entries)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &finance.CreateTransactionResponse{
		TransactionId: txID.String(),
	}, nil
}

func (s *server) GetBalance(ctx context.Context, req *finance.GetBalanceRequest) (*finance.GetBalanceResponse, error) {

	accountID, err := uuid.Parse(req.AccountId)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account id")
	}

	balance, err := s.ledgerService.GetBalance(ctx, accountID)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &finance.GetBalanceResponse{
		Balance: balance.String(),
	}, nil
}

// ===============================
// CREDIT CARD
// ===============================
func (s *server) CreateCreditCard(
	ctx context.Context,
	req *finance.CreateCreditCardRequest,
) (
	*finance.CreateCreditCardResponse,
	error,
) {

	var accountID uuid.UUID
	var err error

	if req.AccountId != "" {
		accountID, err = uuid.Parse(req.AccountId)
		if err != nil {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"invalid accountId: %q",
				req.AccountId,
			)
		}
	}

	limit, err := decimal.NewFromString(req.CreditLimit)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid credit limit",
		)
	}

	cardID, err := s.creditService.CreateCreditCard(
		ctx,
		accountID,
		req.AccountName,
		limit,
		int(req.BillingDay),
		int(req.PaymentDueDays),
	)

	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &finance.CreateCreditCardResponse{
		CreditCardId: cardID.String(),
	}, nil
}

func (s *server) GetAccountSummary(ctx context.Context, req *finance.GetAccountSummaryRequest) (*finance.GetAccountSummaryResponse, error) {

	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account id")
	}

	summary, err := s.ledgerService.GetAccountSummary(ctx, accountID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &finance.GetAccountSummaryResponse{
		AccountId: summary.AccountID.String(),
		Name:      summary.Name,
		Type:      string(summary.Type),
		Balance:   summary.Balance.String(),
		CreatedAt: summary.CreatedAt.Format(time.RFC3339),
	}, nil
}
func (s *server) GetLedgerEntries(ctx context.Context, req *finance.GetLedgerEntriesRequest) (*finance.GetLedgerEntriesResponse, error) {

	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account id")
	}

	entries, err := s.ledgerService.GetLedgerEntries(ctx, accountID)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

func (s *server) GetUserTotalCredit(ctx context.Context, req *finance.GetUserTotalRequest) (*finance.GetUserTotalResponse, error) {

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	total, err := s.ledgerService.GetUserTotalCredit(ctx, userID)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &finance.GetUserTotalResponse{
		UserId:      req.UserId,
		TotalCredit: total.String(),
	}, nil
}

func (s *server) GetUserTotalDebit(ctx context.Context, req *finance.GetUserTotalRequest) (*finance.GetUserTotalResponse, error) {

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	total, err := s.ledgerService.GetUserTotalDebit(ctx, userID)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &finance.GetUserTotalResponse{
		UserId:     req.UserId,
		TotalDebit: total.String(),
	}, nil
}

func (s *server) GetUserTotalBalance(ctx context.Context, req *finance.GetUserTotalRequest) (*finance.GetUserTotalResponse, error) {

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	total, err := s.ledgerService.GetUserTotalBalance(ctx, userID)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &finance.GetUserTotalResponse{
		UserId:       req.UserId,
		TotalBalance: total.String(),
	}, nil
}
func (s *server) RecordCreditCardTransaction(ctx context.Context, req *finance.RecordCreditCardTransactionRequest) (*finance.RecordCreditCardTransactionResponse, error) {

	cardID, err := uuid.Parse(req.CardId)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid card id")
	}

	amount, err := decimal.NewFromString(req.Amount)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}

	txID, err := s.creditService.RecordCreditCardTransaction(ctx, credit.CreditCardTransactionRequest{
		CardID:       cardID,
		Amount:       amount,
		Description:  req.Description,
		PurchaseDate: req.PurchaseDate.AsTime(),
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &finance.RecordCreditCardTransactionResponse{
		TransactionId: txID.String(),
	}, nil
}

func (s *server) PayCreditCardStatement(ctx context.Context, req *finance.PayCreditCardStatementRequest) (*finance.PayCreditCardStatementResponse, error) {

	statementID, err := uuid.Parse(req.StatementId)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid statement id")
	}

	cardID, err := uuid.Parse(req.CardId)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid card id")
	}

	paymentAccountID, err := uuid.Parse(req.PaymentAccountId)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment account id")
	}

	amount, err := decimal.NewFromString(req.Amount)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}

	txID, err := s.creditService.PayCreditCardStatement(ctx, credit.CreditCardPaymentRequest{
		StatementID:      statementID,
		CardID:           cardID,
		PaymentAccountID: paymentAccountID,
		Amount:           amount,
		Description:      req.Description,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &finance.PayCreditCardStatementResponse{
		TransactionId: txID.String(),
	}, nil
}
