package main

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/kaloy/finankal/engine-go/finance"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/credit"
	"github.com/kaloy/finankal/engine-go/internal/db"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/kaloy/finankal/engine-go/internal/loan"
	"github.com/kaloy/finankal/engine-go/internal/schedule"
	"github.com/kaloy/finankal/engine-go/internal/statement"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	loadEnvFromRepoRoot()

	// =====================
	// Infrastructure
	// =====================

	dbConn, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	defer dbConn.Close()

	log.Println("✓ Connected to PostgreSQL")

	redisClient := cache.NewRedis()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	log.Println("✓ Connected to Redis")

	// =====================
	// Ledger
	// =====================

	ledgerRepo := ledger.NewLedgerRepository(dbConn)

	ledgerService := ledger.NewService(
		ledgerRepo,
		redisClient,
	)

	// =====================
	// Credit Card
	// =====================

	cardRepo := credit.NewCardRepository(dbConn)
	statementRepo := credit.NewStatementRepository(dbConn)
	entryRepo := credit.NewEntryRepository(dbConn)

	creditService := credit.NewService(
		cardRepo,
		ledgerRepo,
		statementRepo,
		entryRepo,
		redisClient,
	)

	// =====================
	// Scheduling
	// =====================

	scheduleRepo := schedule.NewScheduleRepository(dbConn, creditService)

	scheduleService := schedule.NewService(
		scheduleRepo,
		ledgerRepo,
		redisClient,
	)

	// =====================
	// gRPC
	// =====================

	server := NewServer(
		ledgerService,
		creditService,
		scheduleService,
		statement.NewService(ledgerRepo, redisClient),
		loan.NewService(dbConn, redisClient),
	)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	finance.RegisterFinanceEngineServer(
		grpcServer,
		server,
	)

	reflection.Register(grpcServer)

	log.Println("✓ Finance Engine running :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func loadEnvFromRepoRoot() {
	// If Docker Compose already provided environment variables,
	// don't bother looking for a .env file.
	if os.Getenv("DATABASE_URL") != "" {
		log.Println("Environment variables already provided. Skipping .env loading.")
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Println("Failed to get working directory:", err)
		return
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				log.Printf("Failed to load .env from %s: %v", envPath, err)
			} else {
				log.Printf("Loaded .env from %s", envPath)
			}
			return
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	log.Println("No .env file found. Using existing environment variables.")
}
