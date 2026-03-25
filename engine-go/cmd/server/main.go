package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/db"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/shopspring/decimal"

	"github.com/google/uuid"

	"github.com/kaloy/finankal/engine-go/internal/model"
	"github.com/kaloy/finankal/engine-go/internal/repository"
)

func main() {
	ctx := context.Background()
	loadEnvFromRepoRoot()
	dbConn, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	redisClient := cache.NewRedis()

	repo := repository.NewLedgerRepository(dbConn)
	service := ledger.NewService(repo, redisClient)
	// 🔥 TEST TRANSACTION
	// 🔹 Replace with actual account IDs from your DB
	cashID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	expenseID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	// 🔹 Create transaction (Cash → Expense)
	entries := []model.Entry{
		{
			AccountID: cashID,
			Amount:    decimal.RequireFromString("500.00"),
			Type:      model.CREDIT,
			CreatedAt: time.Now(),
		},
		{
			AccountID: expenseID,
			Amount:    decimal.RequireFromString("500.00"),
			Type:      model.DEBIT,
			CreatedAt: time.Now(),
		},
	}

	err = service.CreateTransaction(ctx, "Test transaction", entries)
	if err != nil {
		log.Fatal("CreateTransaction error:", err)
	}

	fmt.Println("✅ Transaction created")

	// 🔹 Get balance
	balance, err := service.GetBalance(ctx, cashID)
	if err != nil {
		log.Fatal(err)
	}
	expenseBalance, err := service.GetBalance(ctx, expenseID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("💰 Cash Balance:", balance)
	fmt.Println("💰 Expense Balance:", expenseBalance)

	// 🔹 Get account summary
	summary, err := service.GetAccountSummary(ctx, cashID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("📊 Account Summary:")
	fmt.Println("Name:", summary.Name)
	fmt.Println("Type:", summary.Type)
	fmt.Println("Balance:", summary.Balance)
	fmt.Println("CreatedAt:", summary.CreatedAt)

	summaryExpense, err := service.GetAccountSummary(ctx, expenseID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("📊 Account Summary:")
	fmt.Println("Name:", summaryExpense.Name)
	fmt.Println("Type:", summaryExpense.Type)
	fmt.Println("Balance:", summaryExpense.Balance)
	fmt.Println("CreatedAt:", summaryExpense.CreatedAt)
	
}

func loadEnvFromRepoRoot() {
	dir, err := os.Getwd()
	if err != nil {
		log.Println("Failed to get working directory:", err)
		return
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				log.Println("Failed to load .env:", err)
			} else {
				log.Println(".env loaded from:", envPath)
			}
			return
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	log.Println("No .env file found in repo tree, using system environment variables")
}
