package repository

import (
	"context"
	"fmt"

	"github.com/proyuen/go-mall/internal/model"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id uint) (*model.User, error)
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
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByUsername retrieves a user by their username.
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with username '%s' not found", username)
		}
		return nil, fmt.Errorf("failed to get user by username '%s': %w", username, err)
	}
	return &user, nil
}

// GetByID retrieves a user by their ID.
func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with ID '%d' not found", id)
		}
		return nil, fmt.Errorf("failed to get user by ID '%d': %w", id, err)
	}
	return &user, nil
}