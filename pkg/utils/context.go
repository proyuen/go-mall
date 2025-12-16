package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/pkg/token"
)

const AuthorizationPayloadKey = "authorization_payload"

// GetUserIDFromContext retrieves the UserID from the Gin context.
// It assumes AuthMiddleware has already set the authorization_payload.
func GetUserIDFromContext(c *gin.Context) (uint64, error) {
	payload, exists := c.Get(AuthorizationPayloadKey)
	if !exists {
		return 0, fmt.Errorf("authorization payload not found in context")
	}

	claims, ok := payload.(*token.Payload)
	if !ok {
		return 0, fmt.Errorf("authorization payload is not of type token.Payload")
	}

	return claims.UserID, nil
}
