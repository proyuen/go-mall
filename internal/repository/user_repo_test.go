package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/repository"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createRandomUser is a helper to create a user for testing purposes.
func createRandomUser(t *testing.T, repo repository.UserRepository) *model.User {
	user := &model.User{
		Username:     utils.RandomOwner(),
		PasswordHash: utils.RandomString(32), // Simulate a hash
		Email:        utils.RandomEmail(""),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	if testDB == nil {
		t.Skip("Skipping test because testDB is not initialized")
	}
	// Run in transaction
	tx := testDB.Begin()
	defer tx.Rollback()

	repo := repository.NewUserRepository(tx)
	createRandomUser(t, repo)
}

func TestGetUserByUsername(t *testing.T) {
	if testDB == nil {
		t.Skip("Skipping test because testDB is not initialized")
	}
	// Setup a user to be found
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := repository.NewUserRepository(tx)

	user := createRandomUser(t, repo)

	tests := []struct {
		name          string
		username      string
		expectError   error
		checkResponse func(t *testing.T, retrievedUser *model.User, err error)
	}{
		{
			name:        "Success",
			username:    user.Username,
			expectError: nil,
			checkResponse: func(t *testing.T, retrievedUser *model.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, retrievedUser)
				assert.Equal(t, user.ID, retrievedUser.ID)
				assert.Equal(t, user.Username, retrievedUser.Username)
				assert.Equal(t, user.Email, retrievedUser.Email)
				assert.Equal(t, user.PasswordHash, retrievedUser.PasswordHash)
				assert.WithinDuration(t, user.CreatedAt, retrievedUser.CreatedAt, time.Second)
			},
		},
		{
			name:        "NotFound",
			username:    utils.RandomOwner(), // A username that does not exist
			expectError: repository.ErrUserNotFound,
			checkResponse: func(t *testing.T, retrievedUser *model.User, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, repository.ErrUserNotFound)
				require.Nil(t, retrievedUser)
			},
		},
		{
			name:        "EmptyUsername",
			username:    "",                         // Invalid input
			expectError: repository.ErrUserNotFound, // GORM will return record not found for empty string
			checkResponse: func(t *testing.T, retrievedUser *model.User, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, repository.ErrUserNotFound)
				require.Nil(t, retrievedUser)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			retrievedUser, err := repo.GetByUsername(ctx, tt.username)
			tt.checkResponse(t, retrievedUser, err)
		})
	}
}

func TestGetUserByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Skipping test because testDB is not initialized")
	}
	// Setup a user to be found
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := repository.NewUserRepository(tx)

	user := createRandomUser(t, repo)

	tests := []struct {
		name          string
		id            uint64
		expectError   error
		checkResponse func(t *testing.T, retrievedUser *model.User, err error)
	}{
		{
			name:        "Success",
			id:          user.ID,
			expectError: nil,
			checkResponse: func(t *testing.T, retrievedUser *model.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, retrievedUser)
				assert.Equal(t, user.ID, retrievedUser.ID)
				assert.Equal(t, user.Username, retrievedUser.Username)
				assert.Equal(t, user.Email, retrievedUser.Email)
				assert.Equal(t, user.PasswordHash, retrievedUser.PasswordHash)
				assert.WithinDuration(t, user.CreatedAt, retrievedUser.CreatedAt, time.Second)
			},
		},
		{
			name:        "NotFound",
			id:          999999999999999999, // A very large ID that likely doesn't exist
			expectError: repository.ErrUserNotFound,
			checkResponse: func(t *testing.T, retrievedUser *model.User, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, repository.ErrUserNotFound)
				require.Nil(t, retrievedUser)
			},
		},
		{
			name:        "ZeroID",
			id:          0,                          // Invalid input
			expectError: repository.ErrUserNotFound, // GORM will return record not found for ID 0
			checkResponse: func(t *testing.T, retrievedUser *model.User, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, repository.ErrUserNotFound)
				require.Nil(t, retrievedUser)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			retrievedUser, err := repo.GetByID(ctx, tt.id)
			tt.checkResponse(t, retrievedUser, err)
		})
	}
}
