package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/repository"
	"github.com/proyuen/go-mall/pkg/hasher"
	"github.com/proyuen/go-mall/pkg/token"
)

var (
	ErrUserExists         = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// DTOs (Data Transfer Objects)

type UserRegisterReq struct {
	Username string
	Email    string
	Password string
}

type UserRegisterResp struct {
	UserID   uint64 `json:"user_id,string"` // Snowflake ID
	Username string `json:"username"`
	Email    string `json:"email"`
}

type UserLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLoginResp struct {
	UserID      uint64 `json:"user_id,string"` // Snowflake ID
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"` // Seconds
	TokenType   string `json:"token_type"`
}

//go:generate mockgen -source=$GOFILE -destination=../mocks/user_service_mock.go -package=mocks
// UserService defines the interface for user business logic.
type UserService interface {
	Register(ctx context.Context, req *UserRegisterReq) (*UserRegisterResp, error)
	Login(ctx context.Context, req *UserLoginReq) (*UserLoginResp, error)
}

type userService struct {
	repo       repository.UserRepository
	hasher     hasher.PasswordHasher
	tokenMaker token.Maker
}

// NewUserService creates a new UserService instance.
func NewUserService(repo repository.UserRepository, hasher hasher.PasswordHasher, tokenMaker token.Maker) UserService {
	return &userService{
		repo:       repo,
		hasher:     hasher,
		tokenMaker: tokenMaker,
	}
}

// Register creates a new user.
func (s *userService) Register(ctx context.Context, req *UserRegisterReq) (*UserRegisterResp, error) {
	// 1. Check if user already exists
	// Strict error handling: connection error vs not found error
	_, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		if !errors.Is(err, repository.ErrUserNotFound) {
			// Database connection error or other internal error
			return nil, fmt.Errorf("failed to check existing user: %w", err)
		}
		// ErrUserNotFound means the user does not exist, so we can proceed.
	} else {
		// User found (err == nil), so username is taken
		return nil, ErrUserExists
	}

	// 2. Hash password
	hashedPassword, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	// 3. Create user model
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         "user", // Explicitly set role
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user record: %w", err)
	}

	// 4. Build Response
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 2. Check password
	if err := s.hasher.Check(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 3. Generate Token
	duration := 24 * time.Hour
	accessToken, _, err := s.tokenMaker.CreateToken(user.ID, user.Username, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 4. Build Response
	return &UserLoginResp{
		UserID:      user.ID,
		AccessToken: accessToken,
		ExpiresIn:   int64(duration.Seconds()),
		TokenType:   "Bearer",
	}, nil
}
