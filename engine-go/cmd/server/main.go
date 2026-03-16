package main

import (
	"context"
	_ "fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	// ------------------------
	// 1. Load environment variables
	// ------------------------
	log.Println("Starting server...")
	loadEnvFromRepoRoot()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set in environment variables")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL is not set in environment variables")
	}

	// ------------------------
	// 2. Connect to PostgreSQL using pgx
	// ------------------------
	log.Println("Connecting to PostgreSQL ...", databaseURL)
	pgConn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgConn.Close(context.Background())

	var version string
	if err := pgConn.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		log.Fatalf("Failed to query PostgreSQL: %v", err)
	}
	log.Println("Connected to PostgreSQL:", version)

	// ------------------------
	// 3. Connect to Redis using go-redis
	// ------------------------
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse REDIS_URL: %v", err)
	}
	rdb := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis:", redisURL)

	// ------------------------
	// 4. Start gRPC server
	// ------------------------
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	s := grpc.NewServer()
	// TODO: Register your gRPC services here, e.g.
	// pb.RegisterLedgerServiceServer(s, &ledgerServer{DB: pgConn, Redis: rdb})

	log.Println("gRPC server running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

// loadEnvFromRepoRoot searches up from current directory for .env
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
