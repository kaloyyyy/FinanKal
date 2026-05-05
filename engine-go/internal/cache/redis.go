package cache

import (
	"context"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedis() *redis.Client {
	// Load .env (ignore error if already loaded in main)
	_ = godotenv.Load()

	addr := getEnv("REDIS_ADDR", "localhost:6379")
	password := getEnv("REDIS_PASSWORD", "")
	dbStr := getEnv("REDIS_DB", "0")

	println("Connecting to redis...", addr)
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
