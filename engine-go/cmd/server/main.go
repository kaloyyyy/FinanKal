package main

import (
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	loadEnvFromRepoRoot()

	// =====================
	// Database
	// =====================

	dbConn, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("✓ Connected to PostgreSQL")

	// =====================
	// Redis
	// =====================

	redisClient := cache.NewRedis()

	if err := redisClient.Ping(cache.Ctx).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	log.Println("✓ Connected to Redis")

	// =====================
	// Ledger Module
	// =====================

	ledgerRepo := ledger.NewLedgerRepository(dbConn)

	ledgerService := ledger.NewService(ledgerRepo, redisClient)

	// =====================
	// Credit Card Module
	// =====================

	cardRepo := credit.NewCardRepository(dbConn)

	statementRepo := credit.NewStatementRepository(dbConn)

	entryRepo := credit.NewEntryRepository(dbConn)

	creditService := credit.NewService(dbConn, cardRepo, ledgerRepo, statementRepo, entryRepo, redisClient)

	// =====================
	// gRPC Server
	// =====================

	server := NewServer(ledgerService, creditService)

	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	finance.RegisterFinanceEngineServer(grpcServer, server)

	reflection.Register(grpcServer)

	log.Println("✓ Finance Engine running :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
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
}
