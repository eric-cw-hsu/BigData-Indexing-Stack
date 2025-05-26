package handlers

import (
	"fmt"
	"net/http"

	"eric-cw-hsu.github.io/internal/api/repositories"
	"eric-cw-hsu.github.io/internal/api/services"
	"eric-cw-hsu.github.io/internal/shared/apperror"
	"github.com/gin-gonic/gin"
)

type PlanHandler struct {
	planRepository *repositories.PlanRepository
	planService    *services.PlanService
}

func NewPlanHandler(planRepository *repositories.PlanRepository, planService *services.PlanService) *PlanHandler {
	return &PlanHandler{
		planRepository: planRepository,
		planService:    planService,
	}
}

func (h *PlanHandler) StorePlanHandler(c *gin.Context) {
	var planPayload map[string]interface{}
	if err := c.ShouldBindJSON(&planPayload); err != nil {
		c.JSON(http.StatusBadRequest, apperror.NewInvalidJSONError(err))
		return
	}

	plan, err := h.planService.Create(c, planPayload)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	// create etag for the plan
	etag, err := h.planService.GenerateETag(c, plan)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}
	c.Header("ETag", etag)

	c.JSON(http.StatusOK, gin.H{"message": "Plan stored successfully"})
}

func (h *PlanHandler) GetPlanHandler(c *gin.Context) {
	planId := c.Param("id")

	if h.planService.CheckETag(c, planId, c.GetHeader("If-None-Match")) == nil {
		c.Status(http.StatusNotModified)
		return
	}

	plan, err := h.planService.Get(c, planId)
	if err != nil {
		c.JSON(http.StatusNotFound, apperror.NewPlanNotFoundError(fmt.Errorf("Plan with ID %s not found", planId)))
		return
	}

	// Generate a new ETag for the response
	etag, err := h.planService.GetETag(c, planId)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}
	c.Header("ETag", etag)
	c.JSON(http.StatusOK, plan)
}

func (h *PlanHandler) DeletePlanHandler(c *gin.Context) {
	planId := c.Param("id")

	if err := h.planService.Delete(c, planId); err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	// delete the plan from Redis
	if err := h.planService.DeleteETag(c, planId); err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan deleted successfully"})
}

/*
* updatePlanHandler updates an existing plan.
* The update allow the user to update the plan with a partial json.
* the update plan merged with the exisng plan should be a valid json schema.
 */
func (h *PlanHandler) UpdatePlanHandler(c *gin.Context) {
	planId := c.Param("id")
	var planUpdatePayload map[string]interface{}

	if err := c.ShouldBindBodyWithJSON(&planUpdatePayload); err != nil {
		return
	}

	if c.GetHeader("If-Match") == "" {
		c.JSON(http.StatusPreconditionRequired, apperror.NewETagRequiredError())
		return
	}

	if err := h.planService.CheckETag(c, planId, c.GetHeader("If-Match")); err != nil {
		c.JSON(http.StatusPreconditionFailed, apperror.NewETagNotMatchError())
		return
	}

	plan, err := h.planService.Update(c, planId, planUpdatePayload)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	etag, err := h.planService.GenerateETag(c, plan)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	c.Header("ETag", etag)
	c.JSON(http.StatusOK, gin.H{
		"message": "Plan updated successfully",
		"plan":    plan,
	})
}
