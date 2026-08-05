package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Get retrieves a value from Redis and unmarshals it into the provided type.
// Returns redis.Nil when key does not exist.
func Get[T any](
	ctx context.Context,
	rdb *redis.Client,
	key string,
) (*T, error) {

	if rdb == nil {
		log.Printf(
			"redis disabled: cache get skipped key=%s",
			key,
		)
		return nil, nil
	}

	data, err := rdb.Get(
		ctx,
		key,
	).Bytes()

	if err != nil {

		if IsCacheMiss(err) {
			log.Printf(
				"cache miss key=%s",
				key,
			)
		} else {
			log.Printf(
				"cache get error key=%s err=%v",
				key,
				err,
			)
		}

		return nil, err
	}

	var value T

	err = json.Unmarshal(
		data,
		&value,
	)

	if err != nil {
		log.Printf(
			"cache unmarshal error key=%s err=%v",
			key,
			err,
		)

		return nil, err
	}

	log.Printf(
		"cache hit key=%s",
		key,
	)

	return &value, nil
}

// Set stores a value in Redis with expiration.
func Set[T any](
	ctx context.Context,
	rdb *redis.Client,
	key string,
	value T,
	ttl time.Duration,
) error {

	if rdb == nil {
		log.Printf(
			"redis disabled: cache set skipped key=%s",
			key,
		)

		return nil
	}

	data, err := json.Marshal(
		value,
	)

	if err != nil {
		log.Printf(
			"cache marshal error key=%s err=%v",
			key,
			err,
		)

		return err
	}

	err = rdb.Set(
		ctx,
		key,
		data,
		ttl,
	).Err()

	if err != nil {

		log.Printf(
			"cache set error key=%s ttl=%s err=%v",
			key,
			ttl,
			err,
		)

		return err
	}

	log.Printf(
		"cache set success key=%s ttl=%s",
		key,
		ttl,
	)

	return nil
}

// Exists checks if a cache key exists.
func Exists(
	ctx context.Context,
	rdb *redis.Client,
	key string,
) bool {

	if rdb == nil {
		log.Printf(
			"redis disabled: exists skipped key=%s",
			key,
		)

		return false
	}

	count, err := rdb.Exists(
		ctx,
		key,
	).Result()

	if err != nil {

		log.Printf(
			"cache exists error key=%s err=%v",
			key,
			err,
		)

		return false
	}

	exists := count > 0

	log.Printf(
		"cache exists key=%s result=%v",
		key,
		exists,
	)

	return exists
}

// IsCacheMiss checks if Redis returned a missing key.
func IsCacheMiss(err error) bool {

	return errors.Is(
		err,
		redis.Nil,
	)
}
