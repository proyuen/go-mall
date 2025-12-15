package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test_secret"

	// Generate a valid token for success case
	validToken, err := utils.GenerateToken(1, "testuser", secret)
	require.NoError(t, err)

	type args struct {
		authHeader string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantBody   string
		checkCtx   bool // Whether to check if context keys were set
	}{
		{
			name:       "NoAuthorizationHeader",
			args:       args{authHeader: ""},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "authorization header is not provided",
		},
		{
			name:       "InvalidFormat_NoBearer",
			args:       args{authHeader: "InvalidFormat"},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid authorization header format",
		},
		{
			name:       "UnsupportedAuthorizationType",
			args:       args{authHeader: "Basic token"},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "unsupported authorization type",
		},
		{
			name:       "InvalidToken",
			args:       args{authHeader: "Bearer invalid_token"},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "failed to parse token",
		},
		{
			name:       "ValidToken",
			args:       args{authHeader: "Bearer " + validToken},
			wantStatus: http.StatusOK,
			checkCtx:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/", nil)

			if tt.args.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.args.authHeader)
			}

			// Capture if context was set correctly
			var contextSet bool
			testHandler := func(c *gin.Context) {
				claims, exists := c.Get(authorizationPayloadKey)
				if exists && claims != nil {
					contextSet = true
				}
				c.Status(http.StatusOK)
			}

			// Execute Middleware + Handler
			handler := AuthMiddleware(secret)
			handler(c)
			
			// Only call next handler if not aborted
			if !c.IsAborted() {
				testHandler(c)
			}

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.checkCtx {
				assert.True(t, contextSet, "Context payload should be set for valid token")
			}
		})
	}
}