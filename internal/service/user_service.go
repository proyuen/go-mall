package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/repository"
	"github.com/proyuen/go-mall/pkg/utils"
)

// DTOs (Data Transfer Objects)

type UserRegisterReq struct {
	Username string
	Email    string
	Password string
}

type UserRegisterResp struct {
	UserID   uint
	Username string
	Email    string
}

type UserLoginReq struct {
	Username string
	Password string
}

type UserLoginResp struct {
	AccessToken string
	ExpiresIn   int64 // Seconds
	TokenType   string
}

// UserService defines the interface for user business logic.
type UserService interface {
	Register(ctx context.Context, req *UserRegisterReq) (*UserRegisterResp, error)
	Login(ctx context.Context, req *UserLoginReq) (*UserLoginResp, error)
}

type userService struct {
	repo      repository.UserRepository
	jwtSecret string
}

// NewUserService creates a new UserService instance.
func NewUserService(repo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user.
func (s *userService) Register(ctx context.Context, req *UserRegisterReq) (*UserRegisterResp, error) {
	// 1. Check if user already exists
	existingUser, err := s.repo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 3. Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	// 4. Create user model
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user record: %w", err)
	}

	// 5. Build Response
	return &UserRegisterResp{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// Login authenticates a user and returns a JWT token.
func (s *userService) Login(ctx context.Context, req *UserLoginReq) (*UserLoginResp, error) {
	// 1. Get user
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Log the actual error internally if needed, but return generic error to user
		return nil, errors.New("invalid credentials")
	}

	// 2. Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// 3. Generate Token
	token, err := utils.GenerateToken(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 4. Build Response
	return &UserLoginResp{
		AccessToken: token,
		ExpiresIn:   86400, // 24 hours
		TokenType:   "Bearer",
	}, nil
}