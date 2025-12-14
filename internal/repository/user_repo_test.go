package repository_test

import (
	"context"
	"testing"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/repository"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T, repo repository.UserRepository) *model.User {
	user := &model.User{
		Username:     utils.RandomOwner(),
		PasswordHash: utils.RandomString(32), // Simulate a hash
		Email:        utils.RandomEmail(),
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	// Run in transaction
	tx := testDB.Begin()
	defer tx.Rollback()

	repo := repository.NewUserRepository(tx)
	createRandomUser(t, repo)
}

func TestGetUserByUsername(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := repository.NewUserRepository(tx)

	user1 := createRandomUser(t, repo)

	t.Run("Success", func(t *testing.T) {
		user2, err := repo.GetByUsername(context.Background(), user1.Username)
		require.NoError(t, err)
		require.NotNil(t, user2)

		assert.Equal(t, user1.ID, user2.ID)
		assert.Equal(t, user1.Username, user2.Username)
		assert.Equal(t, user1.Email, user2.Email)
		assert.Equal(t, user1.PasswordHash, user2.PasswordHash)
	})

	t.Run("NotFound", func(t *testing.T) {
		user3, err := repo.GetByUsername(context.Background(), utils.RandomOwner())
		require.Error(t, err)
		require.Nil(t, user3)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestGetUserByID(t *testing.T) {
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := repository.NewUserRepository(tx)

	user1 := createRandomUser(t, repo)

	t.Run("Success", func(t *testing.T) {
		user2, err := repo.GetByID(context.Background(), user1.ID)
		require.NoError(t, err)
		require.NotNil(t, user2)

		assert.Equal(t, user1.ID, user2.ID)
		assert.Equal(t, user1.Username, user2.Username)
	})

	t.Run("NotFound", func(t *testing.T) {
		user3, err := repo.GetByID(context.Background(), 999999) // Assume this ID doesn't exist
		require.Error(t, err)
		require.Nil(t, user3)
		assert.Contains(t, err.Error(), "not found")
	})
}
