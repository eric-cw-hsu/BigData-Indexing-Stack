package handlers

import (
	"fmt"
	"net/http"

	"eric-cw-hsu.github.io/internal/api/repositories"
	"eric-cw-hsu.github.io/internal/api/services"
	"eric-cw-hsu.github.io/internal/api/utils"
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
		c.JSON(http.StatusBadRequest, utils.NewInvalidJSONError(err))
		return
	}

	plan, err := h.planService.Create(c, planPayload)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	// create etag for the plan
	etag := utils.GenerateETag([]byte(fmt.Sprintf("%v", plan)))
	c.Header("ETag", etag)

	c.JSON(http.StatusOK, gin.H{"message": "Plan stored successfully"})
}

func (h *PlanHandler) GetPlanHandler(c *gin.Context) {
	planId := c.Param("id")

	plan, err := h.planService.Get(c, planId)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewPlanNotFoundError(fmt.Errorf("Plan with ID %s not found", planId)))
		return
	}

	// Check if the ETag matches
	if c.GetHeader("If-None-Match") == utils.GenerateETag([]byte(fmt.Sprintf("%v", plan))) {
		c.Status(http.StatusNotModified)
		return
	}

	// Generate a new ETag for the response
	etag := utils.GenerateETag([]byte(fmt.Sprintf("%v", plan)))
	c.Header("ETag", etag)
	c.JSON(http.StatusOK, plan)
}

func (h *PlanHandler) DeletePlanHandler(c *gin.Context) {
	planId := c.Param("id")

	if err := h.planService.Delete(c, planId); err != nil {
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
		c.JSON(http.StatusBadRequest, utils.NewInvalidJSONError(err))
		return
	}

	if c.GetHeader("If-Match") == "" {
		c.JSON(http.StatusPreconditionRequired, utils.NewETagRequiredError())
		return
	}

	plan, err := h.planService.Update(c, planId, planUpdatePayload, c.GetHeader("If-Match"))
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	etag := utils.GenerateETag([]byte(fmt.Sprintf("%v", plan)))
	c.Header("ETag", etag)
	c.JSON(http.StatusOK, gin.H{
		"message": "Plan updated successfully",
		"plan":    plan,
	})
}
