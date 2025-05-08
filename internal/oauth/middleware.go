package oauth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GoogleAuthMiddleware(expectedClientId string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid"})
			c.Abort()
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		googleTokenInfo, err := verifyGoogleToken(token, expectedClientId)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("email", googleTokenInfo.Email)
		c.Set("username", googleTokenInfo.Username)

		c.Next()
	}
}
