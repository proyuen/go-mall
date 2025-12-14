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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockUserService(ctrl)
	handler := NewUserHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))

		mockService.EXPECT().Register(gomock.Any(), gomock.Any()).Return(&service.UserRegisterResp{
			UserID:   1,
			Username: "testuser",
			Email:    "test@example.com",
		}, nil)

		handler.Register(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "User registered successfully")
	})

	t.Run("InvalidInput", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Missing required fields
		reqBody := RegisterRequest{
			Username: "",
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))

		// Service should NOT be called
		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockUserService(ctrl)
	handler := NewUserHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := LoginRequest{
			Username: "testuser",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))

		mockService.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&service.UserLoginResp{
			AccessToken: "mock_token",
		}, nil)

		handler.Login(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "mock_token")
	})

	t.Run("ServiceError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}
		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))

		mockService.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, errors.New("invalid credentials"))

		handler.Login(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
