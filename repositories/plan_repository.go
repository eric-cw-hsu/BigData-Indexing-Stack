package repositories

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type IPlanRepository interface {
	StorePlan(key string, value []byte) error
	GetPlan(key string) (string, error)
	DeletePlan(key string) error
}

type PlanRepository struct {
	redis *redis.Client
}

func NewPlanRepository(redis *redis.Client) *PlanRepository {
	return &PlanRepository{
		redis: redis,
	}
}

func (r *PlanRepository) StorePlan(key string, value []byte) error {
	return r.redis.Set(context.Background(), key, value, 0).Err()
}

func (r *PlanRepository) GetPlan(key string) (string, error) {
	return r.redis.Get(context.Background(), key).Result()
}

func (r *PlanRepository) DeletePlan(key string) error {
	return r.redis.Del(context.Background(), key).Err()
}
