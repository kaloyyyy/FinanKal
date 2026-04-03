package main

import (
	"context"
	"log"

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
	log.Printf("Creating transaction with description: %s", req.Description)
	entries := make([]model.Entry, len(req.Entries))
	accountIDs := make([]string, len(req.Entries))
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
		accountIDs[i] = e.AccountId
	}

	log.Printf("Transaction involves accounts: %v", accountIDs)
	txID, err := s.ledgerService.CreateTransaction(ctx, req.Description, entries)
	if err != nil {
		log.Printf("Failed to create transaction: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}

	log.Printf("Successfully created transaction with ID: %s", txID.String())
	return &finance.CreateTransactionResponse{TransactionId: txID.String()}, nil
}

func (s *server) GetBalance(ctx context.Context, req *finance.GetBalanceRequest) (*finance.GetBalanceResponse, error) {
	log.Printf("Getting balance for account: %s", req.AccountId)
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		log.Printf("Invalid account ID: %s", req.AccountId)
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id: %v", err)
	}

	balance, err := s.ledgerService.GetBalance(ctx, accountID)
	if err != nil {
		log.Printf("Failed to get balance for account %s: %v", req.AccountId, err)
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	log.Printf("Successfully retrieved balance for account %s: %s", req.AccountId, balance.String())
	return &finance.GetBalanceResponse{Balance: balance.String()}, nil
}

func (s *server) GetAccountSummary(ctx context.Context, req *finance.GetAccountSummaryRequest) (*finance.GetAccountSummaryResponse, error) {
	log.Printf("Getting account summary for account: %s", req.AccountId)
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		log.Printf("Invalid account ID: %s", req.AccountId)
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id: %v", err)
	}

	summary, err := s.ledgerService.GetAccountSummary(ctx, accountID)
	if err != nil {
		log.Printf("Failed to get account summary for account %s: %v", req.AccountId, err)
		return nil, status.Errorf(codes.Internal, "failed to get account summary: %v", err)
	}

	log.Printf("Successfully retrieved account summary for account %s", req.AccountId)
	return &finance.GetAccountSummaryResponse{
		AccountId: summary.AccountID.String(),
		Name:      summary.Name,
		Type:      summary.Type,
		Balance:   summary.Balance.String(),
		CreatedAt: summary.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *server) GetLedgerEntries(ctx context.Context, req *finance.GetLedgerEntriesRequest) (*finance.GetLedgerEntriesResponse, error) {
	log.Printf("Getting ledger entries for account: %s", req.AccountId)
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		log.Printf("Invalid account ID: %s", req.AccountId)
		return nil, status.Errorf(codes.InvalidArgument, "invalid account_id: %v", err)
	}

	entries, err := s.ledgerService.GetLedgerEntries(ctx, accountID)
	if err != nil {
		log.Printf("Failed to get ledger entries for account %s: %v", req.AccountId, err)
		return nil, status.Errorf(codes.Internal, "failed to get ledger entries: %v", err)
	}

	log.Printf("Successfully retrieved %d ledger entries for account %s", len(entries), req.AccountId)
	protoEntries := make([]*finance.LedgerEntry, len(entries))
	for i, entry := range entries {
		protoEntries[i] = &finance.LedgerEntry{
			TransactionId: entry.TransactionID.String(),
			AccountId:     entry.AccountID.String(),
			Amount:        entry.Amount.String(),
			Type:          string(entry.Type),
			Description:   entry.Description, // Include transaction description
			CreatedAt:     entry.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &finance.GetLedgerEntriesResponse{Entries: protoEntries}, nil
}
