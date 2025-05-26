package services

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/internal/api/repositories"
	"eric-cw-hsu.github.io/internal/api/schema"
	"eric-cw-hsu.github.io/internal/api/utils"
	"eric-cw-hsu.github.io/internal/objectstore/graph"
	"eric-cw-hsu.github.io/internal/shared/apperror"
	"eric-cw-hsu.github.io/internal/shared/logger"
	"eric-cw-hsu.github.io/internal/shared/messagequeue"
	"eric-cw-hsu.github.io/internal/shared/messagequeue/messages"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type PlanService struct {
	publisher      *messagequeue.Publisher
	planRepository *repositories.PlanRepository
	redisClient    *redis.Client
}

func NewPlanService(publisher *messagequeue.Publisher, planRepository *repositories.PlanRepository, redisClient *redis.Client) *PlanService {
	return &PlanService{
		publisher:      publisher,
		planRepository: planRepository,
		redisClient:    redisClient,
	}
}

func (s *PlanService) publishNodes(nodes map[string]map[string]interface{}, action string) error {
	for _, node := range nodes {
		msg := messages.PlanNodeMessage{
			Action: action,
			Index:  "plans",
			Key:    node["objectId"].(string),
			Data:   node,
		}

		if err := s.publisher.PublishMessage(context.Background(), msg.Type(), msg); err != nil {
			logger.Logger.Error("PlanService.publishNodes: failed to publish message", zap.String("action", action), zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *PlanService) Create(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, *apperror.AppError) {
	if err := schema.ValidateJsonSchema(payload, schema.GetPlanJsonSchema()); err != nil {
		logger.Logger.Error("PlanService.Create: invalid JSON payload", zap.Error(err))
		return nil, apperror.NewInvalidJSONError(err)
	}

	// check if the plan already exists
	planId, _ := payload["objectId"].(string)
	if s.planRepository.IsPlanExists(planId) {
		logger.Logger.Warn("PlanService.Create: plan already exists", zap.String("id", planId))
		return nil, apperror.NewPlanExistsError()
	}

	nodes := graph.ExtractGraphNodes("plan", payload)

	if err := s.planRepository.StorePlanNodes(nodes); err != nil {
		logger.Logger.Error("PlanService.Create: failed to store nodes", zap.Error(err))
		return nil, apperror.NewStorageError("Failed to store plan", err)
	}

	if err := s.publishNodes(nodes, "create"); err != nil {
		logger.Logger.Error("PlanService.Create: publish create failed", zap.Error(err))
		return nil, apperror.NewRabbitMQFailPublishError(err)
	}

	plan, err := s.planRepository.GetPlan(planId)
	if err != nil {
		logger.Logger.Error("PlanService.Create: failed to get plan", zap.Error(err))
		return nil, apperror.NewStorageError("Failed to get plan", err)
	}

	return plan, nil
}

func (s *PlanService) Get(ctx context.Context, id string) (map[string]interface{}, *apperror.AppError) {
	plan, err := s.planRepository.GetPlan(id)
	if err != nil {
		logger.Logger.Error("PlanService.Get: failed to get plan", zap.String("id", id), zap.Error(err))
		return nil, apperror.NewStorageError("Failed to get plan", err)
	}

	return plan, nil
}

func (s *PlanService) Update(
	ctx context.Context,
	id string,
	payload map[string]interface{},
) (map[string]interface{}, *apperror.AppError) {
	if !s.planRepository.IsPlanExists(id) {
		logger.Logger.Warn("PlanService.Update: plan not found", zap.String("id", id))
		return nil, apperror.NewPlanNotFoundError(fmt.Errorf("Plan with ID %s not found", id))
	}

	plan, err := s.planRepository.GetPlan(id)
	if err != nil {
		logger.Logger.Error("PlanService.Update: failed to get existing plan", zap.String("id", id), zap.Error(err))
		return nil, apperror.NewStorageError("Failed to get plan", err)
	}

	mergedPayload, toDeleteObjects, err := graph.Merge(plan, payload)
	if err != nil {
		logger.Logger.Error("PlanService.Update: merge error", zap.Error(err))
		return nil, apperror.NewJSONMergeError(err)
	}

	if err := schema.ValidateJsonSchema(mergedPayload, schema.GetPlanJsonSchema()); err != nil {
		logger.Logger.Error("PlanService.Update: invalid merged JSON", zap.Error(err))
		return nil, apperror.NewInvalidJSONError(err)
	}

	nodes := graph.ExtractGraphNodes("plan", mergedPayload)
	if err := s.planRepository.StorePlanNodes(nodes); err != nil {
		logger.Logger.Error("PlanService.Update: failed to store nodes", zap.Error(err))
		return nil, apperror.NewStorageError("Failed to store plan", err)
	}

	if err := s.publishNodes(nodes, "update"); err != nil {
		logger.Logger.Error("PlanService.Update: publish update failed", zap.Error(err))
		return nil, apperror.NewRabbitMQFailPublishError(err)
	}

	if err := s.publishNodes(toDeleteObjects, "delete"); err != nil {
		logger.Logger.Error("PlanService.Update: publish delete failed", zap.Error(err))
		return nil, apperror.NewRabbitMQFailPublishError(err)
	}

	plan, err = s.planRepository.GetPlan(id)
	if err != nil {
		logger.Logger.Error("PlanService.Update: failed to re-fetch plan", zap.String("id", id), zap.Error(err))
		return nil, apperror.NewStorageError("Failed to get plan", err)
	}

	return plan, nil
}

func (s *PlanService) Delete(ctx context.Context, id string) *apperror.AppError {
	// check if the plan exists
	if !s.planRepository.IsPlanExists(id) {
		logger.Logger.Warn("PlanService.Delete: plan not found", zap.String("id", id))
		return apperror.NewPlanNotFoundError(fmt.Errorf("Plan with ID %s not found", id))
	}

	// delete the plan
	nodes, err := s.planRepository.DeletePlan(id)
	if err != nil {
		logger.Logger.Error("PlanService.Delete: failed to delete plan", zap.String("id", id), zap.Error(err))
		return apperror.NewStorageError("Failed to delete plan", err)
	}

	if err := s.publishNodes(nodes, "delete"); err != nil {
		logger.Logger.Error("PlanService.Delete: publish delete failed", zap.Error(err))
		return apperror.NewRabbitMQFailPublishError(err)
	}

	// delete the plan from Redis
	if err := s.redisClient.Del(ctx, id).Err(); err != nil {
		logger.Logger.Error("PlanService.Delete: failed to delete ETag", zap.String("id", id), zap.Error(err))
		return apperror.NewStorageError("Failed to delete plan from Redis", err)
	}

	return nil
}

func (s *PlanService) GenerateETag(ctx context.Context, plan map[string]interface{}) (string, *apperror.AppError) {
	etag := utils.GenerateETag([]byte(fmt.Sprintf("%v", plan)))
	if err := s.redisClient.Set(ctx, plan["objectId"].(string), etag, 0).Err(); err != nil {
		logger.Logger.Error("PlanService.GenerateETag: failed to set ETag", zap.String("id", plan["objectId"].(string)), zap.Error(err))
		return "", apperror.NewRedisError("Failed to set ETag in Redis", err)
	}

	return etag, nil
}

func (s *PlanService) GetETag(ctx context.Context, id string) (string, *apperror.AppError) {
	etag, err := s.redisClient.Get(ctx, id).Result()
	if err != nil {
		logger.Logger.Warn("PlanService.GetETag: failed to get ETag", zap.String("id", id), zap.Error(err))
		if err == redis.Nil {
			return "", apperror.NewPlanNotFoundError(fmt.Errorf("ETag not found for plan with ID %s", id))
		}

		return "", apperror.NewRedisError("Failed to get ETag from Redis", err)
	}

	return etag, nil
}

func (s *PlanService) CheckETag(ctx context.Context, id string, ifMatch string) *apperror.AppError {
	etag, err := s.GetETag(ctx, id)
	if err != nil {
		return err
	}

	if etag != ifMatch {
		return apperror.NewETagNotMatchError()
	}
	return nil
}

func (s *PlanService) DeleteETag(ctx context.Context, id string) *apperror.AppError {
	err := s.redisClient.Del(ctx, id).Err()
	if err != nil {
		logger.Logger.Error("PlanService.DeleteETag: failed to delete ETag", zap.String("id", id), zap.Error(err))
		return apperror.NewRedisError("Failed to delete ETag from Redis", err)
	}

	return nil
}
