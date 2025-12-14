package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/service"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	userService := service.NewUserService(mockRepo)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &service.UserRegisterReq{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		// Expect GetByUsername to return nil (user not found)
		mockRepo.EXPECT().GetByUsername(ctx, req.Username).Return(nil, nil)

		// Expect Create to be called, and verify the password hash using DoAndReturn
		mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, user *model.User) error {
			assert.True(t, utils.CheckPassword(req.Password, user.PasswordHash), "Password should be hashed correctly")
			return nil // Return nil to simulate successful creation
		})

		resp, err := userService.Register(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Username, resp.Username)
		assert.Equal(t, req.Email, resp.Email)
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		req := &service.UserRegisterReq{
			Username: "existinguser",
			Password: "password123",
		}

		// Expect GetByUsername to return an existing user
		mockRepo.EXPECT().GetByUsername(ctx, req.Username).Return(&model.User{Username: "existinguser"}, nil)

		resp, err := userService.Register(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "username already exists", err.Error())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		req := &service.UserRegisterReq{
			Username: "dbfailuser",
			Email:    "fail@example.com",
			Password: "password123",
		}

		// Expect GetByUsername to return nil (user not found)
		mockRepo.EXPECT().GetByUsername(ctx, req.Username).Return(nil, nil)

		// Expect Create to return an error
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(errors.New("db connection failed"))

		resp, err := userService.Register(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to create user record")
	})
}

func TestUserService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	userService := service.NewUserService(mockRepo)

	ctx := context.Background()

	// Pre-calculate hash for "password123"
	hashedPassword, _ := utils.HashPassword("password123")

	t.Run("Success", func(t *testing.T) {
		req := &service.UserLoginReq{
			Username: "testuser",
			Password: "password123",
		}

		user := &model.User{
			Username:     "testuser",
			PasswordHash: hashedPassword,
		}
		user.ID = 1

		mockRepo.EXPECT().GetByUsername(ctx, req.Username).Return(user, nil)

		resp, err := userService.Login(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		req := &service.UserLoginReq{
			Username: "testuser",
			Password: "wrongpassword",
		}

		user := &model.User{
			Username:     "testuser",
			PasswordHash: hashedPassword,
		}

		mockRepo.EXPECT().GetByUsername(ctx, req.Username).Return(user, nil)

		resp, err := userService.Login(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("UserNotFound", func(t *testing.T) {
		req := &service.UserLoginReq{
			Username: "nonexistent",
			Password: "password123",
		}

		mockRepo.EXPECT().GetByUsername(ctx, req.Username).Return(nil, errors.New("user not found"))

		resp, err := userService.Login(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "invalid credentials", err.Error())
	})
}
