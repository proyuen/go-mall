package repository

import (
	"context"
	"errors" // Add errors import
	"fmt"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/pkg/database"
	"gorm.io/gorm"
)

// ErrUserNotFound is returned when a user record is not found.
var ErrUserNotFound = errors.New("user not found")

//go:generate mockgen -source=$GOFILE -destination=../mocks/user_repo_mock.go -package=mocks
// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id uint64) (*model.User, error) // Changed to uint64
}

// userRepository implements UserRepository using GORM.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create saves a new user to the database.
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	db := database.GetDBFromContext(ctx, r.db)
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByUsername retrieves a user by their username.
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	db := database.GetDBFromContext(ctx, r.db)
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // Use errors.Is for comparison
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username '%s': %w", username, err)
	}
	return &user, nil
}

// GetByID retrieves a user by their ID.
func (r *userRepository) GetByID(ctx context.Context, id uint64) (*model.User, error) { // Changed to uint64
	var user model.User
	db := database.GetDBFromContext(ctx, r.db)
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // Use errors.Is for comparison
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID '%d': %w", id, err)
	}
	return &user, nil
}