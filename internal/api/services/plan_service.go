package services

import (
	"context"
	"encoding/json"
	"fmt"

	"eric-cw-hsu.github.io/internal/api/repositories"
	"eric-cw-hsu.github.io/internal/api/schema"
	"eric-cw-hsu.github.io/internal/api/utils"
	"eric-cw-hsu.github.io/internal/objectstore/graph"
	"eric-cw-hsu.github.io/internal/rabbitmq"
)

type PlanService struct {
	publisher      *rabbitmq.Publisher
	planRepository *repositories.PlanRepository
}

func NewPlanService(publisher *rabbitmq.Publisher, planRepository *repositories.PlanRepository) *PlanService {
	return &PlanService{
		publisher:      publisher,
		planRepository: planRepository,
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
	ifMatch string,
) (map[string]interface{}, *utils.AppError) {
	if !s.planRepository.IsPlanExists(id) {
		return nil, utils.NewPlanNotFoundError(nil)
	}

	plan, err := s.planRepository.GetPlan(id)
	if err != nil {
		return nil, utils.NewStorageError("Failed to get plan", err)
	}

	if utils.GenerateETag([]byte(fmt.Sprintf("%v", plan))) != ifMatch {
		return nil, utils.NewETagNotMatchError()
	}

	mergedPayload, toDeleteObjects, err := graph.Merge(plan, payload)
	if err != nil {
		return nil, utils.NewJSONMergeError(err)
	}

	ori, _ := json.MarshalIndent(plan, "", "  ")
	fmt.Println("Original Payload:", string(ori))

	upd, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Println("Update Payload:", string(upd))

	str, _ := json.MarshalIndent(mergedPayload, "", "  ")
	fmt.Println("Merged Payload:", string(str))

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
		return utils.NewPlanNotFoundError(nil)
	}

	// delete the plan
	nodes, err := s.planRepository.DeletePlan(id)
	if err != nil {
		return utils.NewStorageError("Failed to delete plan", err)
	}

	if err := s.publishNodes(nodes, "delete"); err != nil {
		return utils.NewRabbitMQFailPublishError(err)
	}

	return nil
}
