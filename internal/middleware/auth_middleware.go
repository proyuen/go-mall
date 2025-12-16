package middleware

import (
	"log" // For internal logging
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/pkg/token" // Import the new token package
	"github.com/proyuen/go-mall/pkg/utils"  // For AuthorizationPayloadKey
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
)

// AuthMiddleware creates a Gin middleware for JWT authentication.
// It now takes a token.Maker interface for dependency injection.
func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is not provided"})
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		// Performance: Use EqualFold to avoid memory allocation for ToLower
		if !strings.EqualFold(fields[0], authorizationTypeBearer) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unsupported authorization type"})
			return
		}

		accessToken := fields[1]
		
		// Security: Use tokenMaker.VerifyToken
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			// Security: Do NOT return err.Error() to the client.
			// Log the actual error internally for debugging.
			log.Printf("Failed to verify token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Store payload in context for subsequent handlers
		c.Set(utils.AuthorizationPayloadKey, payload) // Store the actual payload
		c.Next()
	}
}
