package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/internal/service"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Random data
	successUser := utils.RandomOwner()
	successEmail := utils.RandomEmail()
	serviceErrUser := utils.RandomOwner()

	type fields struct {
		mockSetup func(mockService *mocks.MockUserService)
	}
	type args struct {
		reqBody interface{}
	}
	tests := []struct {
		name       string
		args       args
		fields     fields
		wantStatus int
		wantBody   string
	}{
		{
			name: "Success",
			args: args{
				reqBody: RegisterRequest{
					Username: successUser,
					Email:    successEmail,
					Password: "password123",
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockUserService) {
					mockService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(&service.UserRegisterResp{
						UserID:   1,
						Username: successUser,
						Email:    successEmail,
					}, nil)
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "User registered successfully",
		},
		{
			name: "InvalidInput",
			args: args{
				reqBody: RegisterRequest{
					Username: "", // Missing required field
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockUserService) {
					// Service should NOT be called
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Field validation for 'Username' failed on the 'required' tag",
		},
		{
			name: "ServiceError",
			args: args{
				reqBody: RegisterRequest{
					Username: serviceErrUser,
					Email:    utils.RandomEmail(),
					Password: "password123",
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockUserService) {
					mockService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, errors.New("service internal error"))
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "service internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockUserService(ctrl)
			handler := NewUserHandler(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, err := json.Marshal(tt.args.reqBody)
			require.NoError(t, err)

			c.Request, err = http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockService)
			}

			handler.Register(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Random data
	successUser := utils.RandomOwner()
	failUser := utils.RandomOwner()

	type fields struct {
		mockSetup func(mockService *mocks.MockUserService)
	}
	type args struct {
		reqBody interface{}
	}
	tests := []struct {
		name       string
		args       args
		fields     fields
		wantStatus int
		wantBody   string
	}{
		{
			name: "Success",
			args: args{
				reqBody: LoginRequest{
					Username: successUser,
					Password: "password123",
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockUserService) {
					mockService.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&service.UserLoginResp{
						AccessToken: "mock_token",
						ExpiresIn:   86400,
						TokenType:   "Bearer",
					}, nil)
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "mock_token",
		},
		{
			name: "InvalidInput",
			args: args{
				reqBody: LoginRequest{
					Username: "", // Missing field
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockUserService) {
					// Service should NOT be called
				},
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Field validation for 'Username' failed on the 'required' tag",
		},
		{
			name: "AuthenticationFailed",
			args: args{
				reqBody: LoginRequest{
					Username: failUser,
					Password: "wrongpassword",
				},
			},
			fields: fields{
				mockSetup: func(mockService *mocks.MockUserService) {
					mockService.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, errors.New("invalid credentials"))
				},
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockUserService(ctrl)
			handler := NewUserHandler(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, err := json.Marshal(tt.args.reqBody)
			require.NoError(t, err)

			c.Request, err = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockService)
			}

			handler.Login(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}