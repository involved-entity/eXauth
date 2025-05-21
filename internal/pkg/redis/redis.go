package redis

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

func Init(configAddress string, configPassword string, configDB int) {
	redisOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:            configAddress,
			Password:        configPassword,
			DB:              configDB,
			MaxRetries:      3,
			DialTimeout:     5 * time.Second,
			MinRetryBackoff: 300 * time.Millisecond,
			MaxRetryBackoff: 500 * time.Millisecond,
		})

		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			log.Fatal("Failed to connect to Redis: " + err.Error())
		}
	})
}

func GetClient() *redis.Client {
	if redisClient == nil {
		log.Fatal("Redis client not initialized")
	}
	return redisClient
}
