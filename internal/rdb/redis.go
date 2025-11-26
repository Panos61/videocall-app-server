package rdb

import (
	"context"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	once sync.Once
	ctx  context.Context
	rdb  *redis.Client
)

// Initialize the Redis client
func InitRedisClient(url, password string, db int) {
	once.Do(func() {
		ctx = context.Background()
		rdb = redis.NewClient(&redis.Options{
			Addr:     url,
			Password: "",
			DB:       db,
		})

		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}

		log.Println("Connected to Redis successfully.")
	})
}

func Context() context.Context {
	return ctx
}

func Client() *redis.Client {
	return rdb
}
