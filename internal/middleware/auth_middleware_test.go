package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/pkg/token"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Helper function to create a test token maker
func newTestTokenMaker(t *testing.T) token.Maker {
	maker, err := token.NewJWTMaker("12345678901234567890123456789012") // Use a valid-length key
	require.NoError(t, err)
	return maker
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a real token maker for generating valid tokens in test setup
	realTokenMaker := newTestTokenMaker(t)

	// Generate a valid token for success case
	testUserID := uint64(1)
	testUsername := "testuser"
	validToken, validPayload, err := realTokenMaker.CreateToken(testUserID, testUsername, time.Minute)
	require.NoError(t, err)

	type args struct {
		authHeader string
	}
	type fields struct {
		mockSetup func(mockMaker *mocks.MockMaker)
	}
	tests := []struct {
		name       string
		args       args
		fields     fields
		wantStatus int
		wantBody   string
		checkCtx   bool // Whether to check if context payload was set
	}{
		{
			name:       "NoAuthorizationHeader",
			args:       args{authHeader: ""},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "authorization header is not provided",
			checkCtx:   false,
		},
		{
			name:       "InvalidFormat_NoBearer",
			args:       args{authHeader: "InvalidFormat"},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid authorization header format",
			checkCtx:   false,
		},
		{
			name:       "UnsupportedAuthorizationType",
			args:       args{authHeader: "Basic token"},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "unsupported authorization type",
			checkCtx:   false,
		},
		{
			name: "InvalidToken",
			args: args{authHeader: "Bearer invalid_token"},
			fields: fields{
				mockSetup: func(mockMaker *mocks.MockMaker) {
					mockMaker.EXPECT().VerifyToken("invalid_token").Return(nil, token.ErrInvalidToken)
				},
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Unauthorized", // Generic message from middleware
			checkCtx:   false,
		},
		{
			name: "ValidToken",
			args: args{authHeader: "Bearer " + validToken},
			fields: fields{
				mockSetup: func(mockMaker *mocks.MockMaker) {
					// Use gomock.Eq() for string comparison
					mockMaker.EXPECT().VerifyToken(gomock.Eq(validToken)).Return(validPayload, nil)
				},
			},
			wantStatus: http.StatusOK,
			checkCtx:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMaker := mocks.NewMockMaker(ctrl)
			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockMaker)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/", nil)

			if tt.args.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.args.authHeader)
			}

			// Dummy handler to check if context was set
			var contextPayload *token.Payload
			testHandler := func(c *gin.Context) {
				payload, exists := c.Get(utils.AuthorizationPayloadKey)
				if exists {
					contextPayload = payload.(*token.Payload) // Cast to our Payload type
				}
				c.Status(http.StatusOK) // Ensure a status is set if not aborted
			}

			// Execute Middleware
			handler := AuthMiddleware(mockMaker) // Pass mock maker
			handler(c)

			// If middleware didn't abort, call the next handler to test context setting
			if !c.IsAborted() {
				testHandler(c)
			}

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				// Assert response body content for error messages
				var respBody map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &respBody)
				require.NoError(t, err)
				assert.Equal(t, tt.wantBody, respBody["error"])
			}
			if tt.checkCtx {
				require.NotNil(t, contextPayload, "Context payload should be set")
				assert.Equal(t, testUserID, contextPayload.UserID)
				assert.Equal(t, testUsername, contextPayload.Username)
			}
		})
	}
}
