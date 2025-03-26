package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func ErrorHandler(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error(e.Err)
			}
			lastErr := c.Errors.Last().Err
			c.JSON(http.StatusInternalServerError, gin.H{"error": lastErr.Error()})
		}
	}
}
