package repositories

import (
	"eric-cw-hsu.github.io/internal/objectstore/graph"
	"eric-cw-hsu.github.io/internal/objectstore/storage"
	"eric-cw-hsu.github.io/internal/shared/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type PlanRepository struct {
	collection *mongo.Collection
}

func NewPlanRepository(collection *mongo.Collection) *PlanRepository {
	return &PlanRepository{
		collection: collection,
	}
}

func (r *PlanRepository) IsPlanExists(id string) bool {
	_, err := storage.GetNodeRaw(r.collection, id)
	return err == nil
}

func (r *PlanRepository) GetPlan(id string) (map[string]interface{}, error) {
	plan, err := storage.GetExpandedNode(r.collection, id)
	if err != nil {
		logger.Logger.Error("PlanRepository.GetPlan failed", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	return plan, nil
}

func (r *PlanRepository) StorePlanNodes(nodes map[string]map[string]interface{}) error {
	if err := storage.StoreExtractedGraphNodes(r.collection, nodes); err != nil {
		logger.Logger.Error("PlanRepository.StorePlanNodes failed", zap.Error(err))
		return err
	}
	return nil
}

func (r *PlanRepository) DeletePlan(id string) (map[string]map[string]interface{}, error) {
	obj, err := storage.GetExpandedNode(r.collection, id)
	if err != nil {
		logger.Logger.Error("PlanRepository.DeletePlan: fetch failed", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	nodes := graph.ExtractGraphNodes("plan", obj)
	if err := storage.DeleteGraphNodes(r.collection, nodes); err != nil {
		logger.Logger.Error("PlanRepository.DeletePlan: delete graph nodes failed", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return nodes, nil
}
