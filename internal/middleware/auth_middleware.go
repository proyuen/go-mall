package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/pkg/utils"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a Gin middleware for JWT authentication.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
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

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unsupported authorization type"})
			return
		}

		accessToken := fields[1]
		claims, err := utils.ParseToken(accessToken, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Store payload in context for subsequent handlers
		c.Set(authorizationPayloadKey, claims)
		c.Next()
	}
}
