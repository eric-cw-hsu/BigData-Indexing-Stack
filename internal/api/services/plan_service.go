package services

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/internal/api/repositories"
	"eric-cw-hsu.github.io/internal/api/schema"
	"eric-cw-hsu.github.io/internal/api/utils"
	"eric-cw-hsu.github.io/internal/objectstore/graph"
	"eric-cw-hsu.github.io/internal/rabbitmq"
	"github.com/go-redis/redis/v8"
)

type PlanService struct {
	publisher      *rabbitmq.Publisher
	planRepository *repositories.PlanRepository
	redisClient    *redis.Client
}

func NewPlanService(publisher *rabbitmq.Publisher, planRepository *repositories.PlanRepository, redisClient *redis.Client) *PlanService {
	return &PlanService{
		publisher:      publisher,
		planRepository: planRepository,
		redisClient:    redisClient,
	}
}

func (s *PlanService) publishNodes(nodes map[string]map[string]interface{}, action string) error {
	for _, node := range nodes {
		err := s.publisher.Publish(map[string]interface{}{
			"action": action,
			"index":  "plans",
			"key":    node["objectId"],
			"data":   node,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PlanService) Create(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, *utils.AppError) {
	if err := schema.ValidateJsonSchema(payload, schema.GetPlanJsonSchema()); err != nil {
		return nil, utils.NewInvalidJSONError(err)
	}

	// check if the plan already exists
	planId, _ := payload["objectId"].(string)
	if s.planRepository.IsPlanExists(planId) {
		return nil, utils.NewPlanExistsError()
	}

	nodes := graph.ExtractGraphNodes("plan", payload)

	if err := s.planRepository.StorePlanNodes(nodes); err != nil {
		return nil, utils.NewStorageError("Failed to store plan", err)
	}

	if err := s.publishNodes(nodes, "create"); err != nil {
		return nil, utils.NewRabbitMQFailPublishError(err)
	}

	plan, err := s.planRepository.GetPlan(planId)
	if err != nil {
		return nil, utils.NewStorageError("Failed to get plan", err)
	}

	return plan, nil
}

func (s *PlanService) Get(ctx context.Context, id string) (map[string]interface{}, *utils.AppError) {
	plan, err := s.planRepository.GetPlan(id)
	if err != nil {
		return nil, utils.NewStorageError("Failed to get plan", err)
	}

	return plan, nil
}

func (s *PlanService) Update(
	ctx context.Context,
	id string,
	payload map[string]interface{},
) (map[string]interface{}, *utils.AppError) {
	if !s.planRepository.IsPlanExists(id) {
		return nil, utils.NewPlanNotFoundError(fmt.Errorf("Plan with ID %s not found", id))
	}

	plan, err := s.planRepository.GetPlan(id)
	if err != nil {
		return nil, utils.NewStorageError("Failed to get plan", err)
	}

	mergedPayload, toDeleteObjects, err := graph.Merge(plan, payload)
	if err != nil {
		return nil, utils.NewJSONMergeError(err)
	}

	if err := schema.ValidateJsonSchema(mergedPayload, schema.GetPlanJsonSchema()); err != nil {
		return nil, utils.NewInvalidJSONError(err)
	}

	nodes := graph.ExtractGraphNodes("plan", mergedPayload)
	if err := s.planRepository.StorePlanNodes(nodes); err != nil {
		return nil, utils.NewStorageError("Failed to store plan", err)
	}

	if err := s.publishNodes(nodes, "update"); err != nil {
		return nil, utils.NewRabbitMQFailPublishError(err)
	}

	if err := s.publishNodes(toDeleteObjects, "delete"); err != nil {
		return nil, utils.NewRabbitMQFailPublishError(err)
	}

	plan, err = s.planRepository.GetPlan(id)
	if err != nil {
		return nil, utils.NewStorageError("Failed to get plan", err)
	}

	return plan, nil
}

func (s *PlanService) Delete(ctx context.Context, id string) *utils.AppError {
	// check if the plan exists
	if !s.planRepository.IsPlanExists(id) {
		return utils.NewPlanNotFoundError(fmt.Errorf("Plan with ID %s not found", id))
	}

	// delete the plan
	nodes, err := s.planRepository.DeletePlan(id)
	if err != nil {
		return utils.NewStorageError("Failed to delete plan", err)
	}

	if err := s.publishNodes(nodes, "delete"); err != nil {
		return utils.NewRabbitMQFailPublishError(err)
	}

	// delete the plan from Redis
	if err := s.redisClient.Del(ctx, id).Err(); err != nil {
		return utils.NewStorageError("Failed to delete plan from Redis", err)
	}

	return nil
}

func (s *PlanService) GenerateETag(ctx context.Context, plan map[string]interface{}) (string, *utils.AppError) {
	etag := utils.GenerateETag([]byte(fmt.Sprintf("%v", plan)))
	if err := s.redisClient.Set(ctx, plan["objectId"].(string), etag, 0).Err(); err != nil {
		return "", utils.NewRedisError("Failed to set ETag in Redis", err)
	}

	return etag, nil
}

func (s *PlanService) GetETag(ctx context.Context, id string) (string, *utils.AppError) {
	etag, err := s.redisClient.Get(ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			return "", utils.NewPlanNotFoundError(nil)
		}

		return "", utils.NewRedisError("Failed to get ETag from Redis", err)
	}

	return etag, nil
}

func (s *PlanService) CheckETag(ctx context.Context, id string, ifMatch string) *utils.AppError {
	etag, err := s.GetETag(ctx, id)
	if err != nil {
		return err
	}

	if etag != ifMatch {
		return utils.NewETagNotMatchError()
	}
	return nil
}

func (s *PlanService) DeleteETag(ctx context.Context, id string) *utils.AppError {
	err := s.redisClient.Del(ctx, id).Err()
	if err != nil {
		return utils.NewRedisError("Failed to delete ETag from Redis", err)
	}

	return nil
}
