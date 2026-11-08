package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetOrLoad returns the cached value for key, or calls load and caches its result.
// Cache failures are logged and never fail the request. A nil client disables caching.
func GetOrLoad[T any](
	ctx context.Context,
	rdb *redis.Client,
	key string,
	ttl time.Duration,
	load func() (T, error),
) (T, error) {

	if rdb == nil {
		return load()
	}

	data, err := rdb.Get(ctx, key).Bytes()
	if err == nil {
		var value T
		err := json.Unmarshal(data, &value)
		if err == nil {
			return value, nil
		}
		log.Printf("cache unmarshal error key=%s err=%v", key, err)
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("cache get error key=%s err=%v", key, err)
	}

	value, err := load()
	if err != nil {
		return value, err
	}

	data, err = json.Marshal(value)
	if err != nil {
		log.Printf("cache marshal error key=%s err=%v", key, err)
		return value, nil
	}

	if err := rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("cache set error key=%s err=%v", key, err)
	}

	return value, nil
}
