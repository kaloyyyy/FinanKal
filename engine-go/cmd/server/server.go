package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/kaloy/finankal/engine-go/finance"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/kaloy/finankal/engine-go/internal/model"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	finance.UnimplementedFinanceEngineServer
	ledgerService *ledger.Service
}

func NewServer(ledgerService *ledger.Service) *server {
	return &server{
		ledgerService: ledgerService,
	}
}

func (s *server) HealthCheck(ctx context.Context, req *finance.HealthRequest) (*finance.HealthResponse, error) {
	return &finance.HealthResponse{Status: "OK"}, nil
}

func (s *server) CreateTransaction(ctx context.Context, req *finance.CreateTransactionRequest) (*finance.CreateTransactionResponse, error) {
	entries := make([]model.Entry, len(req.Entries))
	for i, e := range req.Entries {
		accountID, err := uuid.Parse(e.AccountId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid account_id: %v", err)
		}
		amount, err := decimal.NewFromString(e.Amount)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid amount: %v", err)
		}
		var entryType model.EntryType
		if e.Type == "DEBIT" {
			entryType = model.DEBIT
		} else if e.Type == "CREDIT" {
			entryType = model.CREDIT
		} else {
			return nil, status.Errorf(codes.InvalidArgument, "invalid type: must be DEBIT or CREDIT")
		}
		entries[i] = model.Entry{
			AccountID: accountID,
			Amount:    amount,
			Type:      entryType,
		}
	}

	txID, err := s.ledgerService.CreateTransaction(ctx, req.Description, entries)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}

	return &finance.CreateTransactionResponse{TransactionId: txID.String()}, nil
}

func (s *server) GetBalance(ctx context.Context, req *finance.GetBalanceRequest) (*finance.GetBalanceResponse, error) {
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id: %v", err)
	}

	balance, err := s.ledgerService.GetBalance(ctx, accountID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	return &finance.GetBalanceResponse{Balance: balance.String()}, nil
}

func (s *server) GetAccountSummary(ctx context.Context, req *finance.GetAccountSummaryRequest) (*finance.GetAccountSummaryResponse, error) {
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id: %v", err)
	}

	summary, err := s.ledgerService.GetAccountSummary(ctx, accountID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get account summary: %v", err)
	}

	return &finance.GetAccountSummaryResponse{
		AccountId: summary.AccountID.String(),
		Name:      summary.Name,
		Type:      summary.Type,
		Balance:   summary.Balance.String(),
		CreatedAt: summary.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
