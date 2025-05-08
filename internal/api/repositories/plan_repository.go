package repositories

import (
	"eric-cw-hsu.github.io/internal/objectstore/graph"
	"eric-cw-hsu.github.io/internal/objectstore/storage"
	"go.mongodb.org/mongo-driver/mongo"
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
	return storage.GetExpandedNode(r.collection, id)
}

func (r *PlanRepository) StorePlanNodes(nodes map[string]map[string]interface{}) error {
	return storage.StoreExtractedGraphNodes(r.collection, nodes)
}

func (r *PlanRepository) DeletePlan(id string) (map[string]map[string]interface{}, error) {
	obj, err := storage.GetExpandedNode(r.collection, id)
	if err != nil {
		return nil, err
	}

	nodes := graph.ExtractGraphNodes("plan", obj)
	if err := storage.DeleteGraphNodes(r.collection, nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}
