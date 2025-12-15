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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserService_Register(t *testing.T) {
	// Generate random data for test cases
	successUser := utils.RandomOwner()
	successEmail := utils.RandomEmail()
	existingUser := utils.RandomOwner()
	dbFailUser := utils.RandomOwner()
	dbFailEmail := utils.RandomEmail()

	type fields struct {
		mockSetup func(mockRepo *mocks.MockUserRepository, req *service.UserRegisterReq)
	}
	type args struct {
		req *service.UserRegisterReq
	}
	tests := []struct {
		name      string
		args      args
		fields    fields
		wantErr   bool
		errStr    string
		wantResp  bool
		checkResp func(t *testing.T, resp *service.UserRegisterResp, req *service.UserRegisterReq)
	}{
		{
			name: "Success",
			args: args{
				req: &service.UserRegisterReq{
					Username: successUser,
					Email:    successEmail,
					Password: "password123",
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockUserRepository, req *service.UserRegisterReq) {
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(nil, errors.New("user not found"))
					mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, user *model.User) error {
						require.True(t, utils.CheckPassword(req.Password, user.PasswordHash), "Password should be hashed correctly")
						return nil
					})
				},
			},
			wantErr:  false,
			wantResp: true,
			checkResp: func(t *testing.T, resp *service.UserRegisterResp, req *service.UserRegisterReq) {
				assert.Equal(t, req.Username, resp.Username)
				assert.Equal(t, req.Email, resp.Email)
			},
		},
		{
			name: "UserAlreadyExists",
			args: args{
				req: &service.UserRegisterReq{
					Username: existingUser,
					Password: "password123",
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockUserRepository, req *service.UserRegisterReq) {
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(&model.User{Username: existingUser}, nil)
				},
			},
			wantErr: true,
			errStr:  "username already exists",
		},
		{
			name: "DatabaseError_Create",
			args: args{
				req: &service.UserRegisterReq{
					Username: dbFailUser,
					Email:    dbFailEmail,
					Password: "password123",
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockUserRepository, req *service.UserRegisterReq) {
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(nil, errors.New("user not found"))
					mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db connection failed"))
				},
			},
			wantErr: true,
			errStr:  "failed to create user record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUserRepository(ctrl)
			userService := service.NewUserService(mockRepo, "test_secret")
			ctx := context.Background()

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockRepo, tt.args.req)
			}

			resp, err := userService.Register(ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errStr != "" {
					assert.Contains(t, err.Error(), tt.errStr)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.wantResp {
				require.NotNil(t, resp)
				if tt.checkResp != nil {
					tt.checkResp(t, resp, tt.args.req)
				}
			} else {
				require.Nil(t, resp)
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("password123")
	successUser := utils.RandomOwner()
	notFoundUser := utils.RandomOwner()

	type fields struct {
		mockSetup func(mockRepo *mocks.MockUserRepository, req *service.UserLoginReq)
	}
	type args struct {
		req *service.UserLoginReq
	}
	tests := []struct {
		name     string
		args     args
		fields   fields
		wantErr  bool
		errStr   string
		wantResp bool
	}{
		{
			name: "Success",
			args: args{
				req: &service.UserLoginReq{
					Username: successUser,
					Password: "password123",
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockUserRepository, req *service.UserLoginReq) {
					user := &model.User{
						Username:     successUser,
						PasswordHash: hashedPassword,
					}
					user.ID = 1
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(user, nil)
				},
			},
			wantErr:  false,
			wantResp: true,
		},
		{
			name: "InvalidPassword",
			args: args{
				req: &service.UserLoginReq{
					Username: successUser,
					Password: "wrongpassword",
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockUserRepository, req *service.UserLoginReq) {
					user := &model.User{
						Username:     successUser,
						PasswordHash: hashedPassword,
					}
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(user, nil)
				},
			},
			wantErr: true,
			errStr:  "invalid credentials",
		},
		{
			name: "UserNotFound",
			args: args{
				req: &service.UserLoginReq{
					Username: notFoundUser,
					Password: "password123",
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockUserRepository, req *service.UserLoginReq) {
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(nil, errors.New("user not found"))
				},
			},
			wantErr: true,
			errStr:  "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUserRepository(ctrl)
			userService := service.NewUserService(mockRepo, "test_secret")
			ctx := context.Background()

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockRepo, tt.args.req)
			}

			resp, err := userService.Login(ctx, tt.args.req)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errStr != "" {
					assert.Contains(t, err.Error(), tt.errStr)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.wantResp {
				require.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
			} else {
				require.Nil(t, resp)
			}
		})
	}
}