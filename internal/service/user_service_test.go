package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/proyuen/go-mall/internal/mocks"
	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/repository"
	"github.com/proyuen/go-mall/internal/service"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserService_Register(t *testing.T) {
	// Generate random data for test cases
	successUser := utils.RandomOwner()
	successEmail := utils.RandomEmail("")
	existingUser := utils.RandomOwner()
	dbFailUser := utils.RandomOwner()
	dbFailEmail := utils.RandomEmail("")

	type fields struct {
		mockSetup func(mockRepo *mocks.MockUserRepository, mockHasher *mocks.MockPasswordHasher, mockMaker *mocks.MockMaker, req *service.UserRegisterReq)
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
				mockSetup: func(mockRepo *mocks.MockUserRepository, mockHasher *mocks.MockPasswordHasher, mockMaker *mocks.MockMaker, req *service.UserRegisterReq) {
					// Expect user check -> returns Not Found (good for registration)
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(nil, repository.ErrUserNotFound)
					
					// Expect password hashing
					hashedPassword := "hashed_password_123"
					mockHasher.EXPECT().Hash(req.Password).Return(hashedPassword, nil)

					mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, user *model.User) error {
						// Verify the hashed password was used
						assert.Equal(t, hashedPassword, user.PasswordHash)
						// Simulate ID generation
						user.ID = 101
						return nil
					})
				},
			},
			wantErr:  false,
			wantResp: true,
			checkResp: func(t *testing.T, resp *service.UserRegisterResp, req *service.UserRegisterReq) {
				assert.Equal(t, req.Username, resp.Username)
				assert.Equal(t, req.Email, resp.Email)
				assert.Equal(t, uint64(101), resp.UserID)
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
				mockSetup: func(mockRepo *mocks.MockUserRepository, mockHasher *mocks.MockPasswordHasher, mockMaker *mocks.MockMaker, req *service.UserRegisterReq) {
					// Expect user check -> returns User (bad for registration)
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(&model.User{Username: existingUser}, nil)
				},
			},
			wantErr: true,
			errStr:  "username already exists",
		},
		{
			name: "DatabaseError_CheckUser",
			args: args{
				req: &service.UserRegisterReq{
					Username: dbFailUser,
					Email:    dbFailEmail,
					Password: "password123",
				},
			},
			fields: fields{
				mockSetup: func(mockRepo *mocks.MockUserRepository, mockHasher *mocks.MockPasswordHasher, mockMaker *mocks.MockMaker, req *service.UserRegisterReq) {
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(nil, errors.New("db connection failed"))
				},
			},
			wantErr: true,
			errStr:  "failed to check existing user",
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
				mockSetup: func(mockRepo *mocks.MockUserRepository, mockHasher *mocks.MockPasswordHasher, mockMaker *mocks.MockMaker, req *service.UserRegisterReq) {
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(nil, repository.ErrUserNotFound)
					
					hashedPassword := "hashed_password_123"
					mockHasher.EXPECT().Hash(req.Password).Return(hashedPassword, nil)
					
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
			mockHasher := mocks.NewMockPasswordHasher(ctrl)
			mockMaker := mocks.NewMockMaker(ctrl)
			
			userService := service.NewUserService(mockRepo, mockHasher, mockMaker)
			ctx := context.Background()

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockRepo, mockHasher, mockMaker, tt.args.req)
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
	hashedPassword := "mock_hashed_password"
	successUser := utils.RandomOwner()
	notFoundUser := utils.RandomOwner()

	type fields struct {
		mockSetup func(mockRepo *mocks.MockUserRepository, mockHasher *mocks.MockPasswordHasher, mockMaker *mocks.MockMaker, req *service.UserLoginReq)
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
				mockSetup: func(mockRepo *mocks.MockUserRepository, mockHasher *mocks.MockPasswordHasher, mockMaker *mocks.MockMaker, req *service.UserLoginReq) {
					user := &model.User{
						Username:     successUser,
						PasswordHash: hashedPassword,
					}
					user.ID = 101 // uint64
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(user, nil)
					
					// Expect password check
					mockHasher.EXPECT().Check(req.Password, hashedPassword).Return(nil)

					// Expect token generation
					mockMaker.EXPECT().CreateToken(user.ID, user.Username, 24*time.Hour).Return("mock_access_token", nil, nil)
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
				mockSetup: func(mockRepo *mocks.MockUserRepository, mockHasher *mocks.MockPasswordHasher, mockMaker *mocks.MockMaker, req *service.UserLoginReq) {
					user := &model.User{
						Username:     successUser,
						PasswordHash: hashedPassword,
					}
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(user, nil)
					
					// Expect password check failure
					mockHasher.EXPECT().Check(req.Password, hashedPassword).Return(errors.New("invalid password"))
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
				mockSetup: func(mockRepo *mocks.MockUserRepository, mockHasher *mocks.MockPasswordHasher, mockMaker *mocks.MockMaker, req *service.UserLoginReq) {
					mockRepo.EXPECT().GetByUsername(gomock.Any(), req.Username).Return(nil, repository.ErrUserNotFound)
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
			mockHasher := mocks.NewMockPasswordHasher(ctrl)
			mockMaker := mocks.NewMockMaker(ctrl)

			userService := service.NewUserService(mockRepo, mockHasher, mockMaker)
			ctx := context.Background()

			if tt.fields.mockSetup != nil {
				tt.fields.mockSetup(mockRepo, mockHasher, mockMaker, tt.args.req)
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
				// assert.Equal(t, uint64(101), resp.UserID)
			} else {
				require.Nil(t, resp)
			}
		})
	}
}
