package services

import (
	"encoding/json"
	"fmt"

	"eric-cw-hsu.github.io/errors"
	"eric-cw-hsu.github.io/models"
	"eric-cw-hsu.github.io/queue"
	"eric-cw-hsu.github.io/repositories"
	"eric-cw-hsu.github.io/utils"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type PlanService struct {
	planRepository repositories.IPlanRepository
	logger         *logrus.Logger
	rabbitQueue    *queue.RabbitMQQueue
}

// NewPlanService now accepts a RabbitMQQueue instance
func NewPlanService(planRepository repositories.IPlanRepository, logger *logrus.Logger, rabbitQueue *queue.RabbitMQQueue) *PlanService {
	return &PlanService{
		planRepository: planRepository,
		logger:         logger,
		rabbitQueue:    rabbitQueue,
	}
}

func (s *PlanService) CreatePlan(payload map[string]interface{}) (string, string, error) {
	err := utils.ValidateJSONSchema(payload, models.GetPlanJsonSchema())
	if err != nil {
		s.logger.Error("JSON Schema validation error:", err)
		return "", "", err
	}

	key := payload["objectType"].(string) + ":" + payload["objectId"].(string)
	if _, err := s.planRepository.GetPlan(key); err != redis.Nil {
		return "", "", errors.ErrKeyAlreadyExists
	}

	dataBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("JSON Marshal error:", err)
		return "", "", err
	}

	if err := s.planRepository.StorePlan(key, dataBytes); err != nil {
		s.logger.Error("StorePlan error:", err)
		return "", "", err
	}

	etag := fmt.Sprintf("%x", utils.GenerateETag(dataBytes))

	// Push creation event with the complete document to queue
	if s.rabbitQueue != nil {
		createEvent := map[string]interface{}{
			"action": "create",
			"key":    key,
			"data":   payload,
		}
		eventBytes, err := json.Marshal(createEvent)
		if err != nil {
			s.logger.Error("Failed to marshal create event:", err)
		} else {
			if pubErr := s.rabbitQueue.Publish(eventBytes); pubErr != nil {
				s.logger.Error("Failed to publish create event:", pubErr)
			}
		}
	} else {
		s.logger.Warn("RabbitMQQueue not available, skipping event publishing for create action")
	}

	return key, etag, nil
}

func (s *PlanService) GetPlan(key string) (string, string, error) {
	val, err := s.planRepository.GetPlan(key)
	if err == redis.Nil {
		return "", "", errors.ErrKeyNotFound
	} else if err != nil {
		s.logger.Error("GetPlan error:", err)
		return "", "", err
	}

	dataBytes := []byte(val)
	etag := fmt.Sprintf("%x", utils.GenerateETag(dataBytes))
	return val, etag, nil
}

func (s *PlanService) DeletePlan(key string) error {
	if _, err := s.planRepository.GetPlan(key); err == redis.Nil {
		return errors.ErrKeyNotFound
	} else if err != nil {
		s.logger.Error("GetPlan error in DeletePlan:", err)
		return err
	}

	if err := s.planRepository.DeletePlan(key); err != nil {
		s.logger.Error("DeletePlan error:", err)
		return err
	}

	// Push deletion event to queue (key is sufficient)
	if s.rabbitQueue != nil {
		deleteEvent := map[string]interface{}{
			"action": "delete",
			"key":    key,
		}
		eventBytes, err := json.Marshal(deleteEvent)
		if err != nil {
			s.logger.Error("Failed to marshal delete event:", err)
		} else {
			if pubErr := s.rabbitQueue.Publish(eventBytes); pubErr != nil {
				s.logger.Error("Failed to publish delete event:", pubErr)
			}
		}
	} else {
		s.logger.Warn("RabbitMQQueue not available, skipping event publishing for delete action")
	}

	return nil
}

func (s *PlanService) UpdatePlan(key string, payload map[string]interface{}) (string, string, error) {
	val, err := s.planRepository.GetPlan(key)
	if err == redis.Nil {
		return "", "", errors.ErrKeyNotFound
	} else if err != nil {
		s.logger.Error("GetPlan error in UpdatePlan:", err)
		return "", "", err
	}

	existingData := make(map[string]interface{})
	if err := json.Unmarshal([]byte(val), &existingData); err != nil {
		s.logger.Error("Unmarshal error in UpdatePlan:", err)
		return "", "", err
	}

	mergedData := mergePlans(existingData, payload)

	if err := utils.ValidateJSONSchema(mergedData, models.GetPlanJsonSchema()); err != nil {
		s.logger.Error("JSON Schema validation error:", err)
		return "", "", err
	}

	dataBytes, err := json.Marshal(mergedData)
	if err != nil {
		s.logger.Error("JSON Marshal error in UpdatePlan:", err)
		return "", "", err
	}

	if err := s.planRepository.StorePlan(key, dataBytes); err != nil {
		s.logger.Error("StorePlan error in UpdatePlan:", err)
		return "", "", err
	}

	etag := fmt.Sprintf("%x", utils.GenerateETag(dataBytes))

	// Push update event with the complete updated document to queue
	if s.rabbitQueue != nil {
		updateEvent := map[string]interface{}{
			"action": "update",
			"key":    key,
			"data":   mergedData,
		}
		eventBytes, err := json.Marshal(updateEvent)
		if err != nil {
			s.logger.Error("Failed to marshal update event:", err)
		} else {
			if pubErr := s.rabbitQueue.Publish(eventBytes); pubErr != nil {
				s.logger.Error("Failed to publish update event:", pubErr)
			}
		}
	} else {
		s.logger.Warn("RabbitMQQueue not available, skipping event publishing for update action")
	}

	return string(dataBytes), etag, nil
}

func mergePlans(existing, newPlan map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range newPlan {
		if k == "linkedPlanServices" {
			merged[k] = mergeLinkedPlanServices(existing[k], v)
		} else {
			merged[k] = v
		}
	}
	return merged
}

func mergeLinkedPlanServices(existing, newData interface{}) []map[string]interface{} {
	existingList, ok1 := existing.([]interface{})
	newList, ok2 := newData.([]interface{})
	if !ok1 {
		existingList = []interface{}{}
	}
	if !ok2 {
		newList = []interface{}{}
	}

	existingMap := make(map[string]map[string]interface{})
	for _, item := range existingList {
		if obj, ok := item.(map[string]interface{}); ok {
			if id, exists := obj["objectId"].(string); exists {
				existingMap[id] = obj
			}
		}
	}
	for _, item := range newList {
		if obj, ok := item.(map[string]interface{}); ok {
			if id, exists := obj["objectId"].(string); exists {
				existingMap[id] = obj
			}
		}
	}
	mergedList := []map[string]interface{}{}
	for _, obj := range existingMap {
		mergedList = append(mergedList, obj)
	}
	return mergedList
}
