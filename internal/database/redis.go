package database

import (
	"log"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(uri string) *RedisService {
	redisClient := redis.NewClient(&redis.Options{
		Addr: uri,
	})

	return &RedisService{
		client: redisClient,
	}
}

func (r *RedisService) GetClient() *redis.Client {
	return r.client
}

func (r *RedisService) Close() {
	if err := r.client.Close(); err != nil {
		log.Printf("Failed to disconnect from Redis: %v", err)
	}
}
