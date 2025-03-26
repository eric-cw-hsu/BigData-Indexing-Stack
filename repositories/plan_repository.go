package repositories

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type IPlanRepository interface {
	StorePlan(key string, value []byte) error
	GetPlan(key string) (string, error)
	DeletePlan(key string) error
}

type PlanRepository struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewPlanRepository(redis *redis.Client, logger *logrus.Logger) *PlanRepository {
	return &PlanRepository{
		redis:  redis,
		logger: logger,
	}
}

func (r *PlanRepository) StorePlan(key string, value []byte) error {
	err := r.redis.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		r.logger.WithField("key", key).Error("Failed to store plan:", err)
		return err
	}
	r.logger.WithField("key", key).Info("Plan stored successfully")
	return nil
}

func (r *PlanRepository) GetPlan(key string) (string, error) {
	result, err := r.redis.Get(context.Background(), key).Result()
	if err != nil {
		r.logger.WithField("key", key).Error("Failed to get plan:", err)
		return "", err
	}
	r.logger.WithField("key", key).Info("Plan retrieved successfully")
	return result, nil
}

func (r *PlanRepository) DeletePlan(key string) error {
	err := r.redis.Del(context.Background(), key).Err()
	if err != nil {
		r.logger.WithField("key", key).Error("Failed to delete plan:", err)
		return err
	}
	r.logger.WithField("key", key).Info("Plan deleted successfully")
	return nil
}
