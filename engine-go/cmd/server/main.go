package main

import (
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/kaloy/finankal/engine-go/finance"
	"github.com/kaloy/finankal/engine-go/internal/cache"
	"github.com/kaloy/finankal/engine-go/internal/db"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"github.com/kaloy/finankal/engine-go/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	loadEnvFromRepoRoot()
	dbConn, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	redisClient := cache.NewRedis()

	repo := repository.NewLedgerRepository(dbConn)
	ledgerService := ledger.NewService(repo, redisClient)

	financeServer := NewServer(ledgerService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	finance.RegisterFinanceEngineServer(s, financeServer)
	reflection.Register(s)

	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
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
