package database

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/config"
)

func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.GetEnv("REDIS_HOST"), config.GetEnv("REDIS_PORT")),
		Password: config.GetEnv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Test the connection
	if _, err := client.Ping().Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis!")
	return client, nil
}
