package repository_test

import (
	"context"
	"testing"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/repository"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCategoryID      = uint64(1)
	testPaginationLimit = 5
	testSPUCount        = 10
	nonExistentID       = uint64(999999999)
)

// createRandomSPU creates a random SPU with associated SKUs for testing.
// It returns the created SPU or an error, allowing the caller to handle assertions.
func createRandomSPU(ctx context.Context, repo repository.ProductRepository) (*model.SPU, error) {
	// Use structured map for attributes to avoid JSON injection risks
	attr1 := model.JSONB{
		"color": utils.RandomString(5),
		"size":  "M",
	}
	attr2 := model.JSONB{
		"color": utils.RandomString(5),
		"size":  "L",
	}

	spu := &model.SPU{
		Name:        utils.RandomString(10),
		Description: utils.RandomString(20),
		CategoryID:  testCategoryID, // Use constant to avoid flaky tests with random invalid IDs
		SKUs: []model.SKU{
			{
				Attributes: attr1,
				Price:      decimal.NewFromInt(int64(utils.RandomInt(10, 100))),
				Stock:      int(utils.RandomInt(1, 100)),
			},
			{
				Attributes: attr2,
				Price:      decimal.NewFromInt(int64(utils.RandomInt(10, 100))),
				Stock:      int(utils.RandomInt(1, 100)),
			},
		},
	}

	err := repo.CreateSPU(ctx, spu)
	if err != nil {
		return nil, err
	}
	return spu, nil
}

func TestCreateSPU(t *testing.T) {
	if testDB == nil {
		t.Skip("Skipping test because testDB is not initialized")
	}
	tx := testDB.Begin()
	defer tx.Rollback()

	repo := repository.NewProductRepository(tx)
	ctx := context.Background()

	spu, err := createRandomSPU(ctx, repo)
	require.NoError(t, err)
	require.NotNil(t, spu)
	require.NotZero(t, spu.ID)
	require.NotZero(t, spu.CreatedAt)

	for _, sku := range spu.SKUs {
		require.NotZero(t, sku.ID)
		require.Equal(t, spu.ID, sku.SPUID)
	}
}

func TestGetSPUByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Skipping test because testDB is not initialized")
	}
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := repository.NewProductRepository(tx)
	ctx := context.Background()

	spu1, err := createRandomSPU(ctx, repo)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		spu2, err := repo.GetSPUByID(ctx, spu1.ID)
		require.NoError(t, err)
		require.NotNil(t, spu2)

		assert.Equal(t, spu1.ID, spu2.ID)
		assert.Equal(t, spu1.Name, spu2.Name)
	})

	t.Run("NotFound", func(t *testing.T) {
		spu3, err := repo.GetSPUByID(ctx, nonExistentID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrSPUNotFound) // Assert sentinel error
		require.Nil(t, spu3)
	})
}

func TestListSPUs(t *testing.T) {
	if testDB == nil {
		t.Skip("Skipping test because testDB is not initialized")
	}
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := repository.NewProductRepository(tx)
	ctx := context.Background()

	// Create multiple SPUs
	for i := 0; i < testSPUCount; i++ {
		_, err := createRandomSPU(ctx, repo)
		require.NoError(t, err)
	}

	t.Run("DefaultPagination", func(t *testing.T) {
		spus, err := repo.ListSPUs(ctx, 0, testPaginationLimit)
		require.NoError(t, err)
		assert.Len(t, spus, testPaginationLimit)
	})

	t.Run("OffsetPagination", func(t *testing.T) {
		spus, err := repo.ListSPUs(ctx, testPaginationLimit, testPaginationLimit)
		require.NoError(t, err)
		assert.Len(t, spus, testPaginationLimit)
	})

	t.Run("NoResults", func(t *testing.T) {
		spus, err := repo.ListSPUs(ctx, 1000, testPaginationLimit) // Large offset
		require.NoError(t, err)
		assert.Empty(t, spus)
	})
}

func TestGetSKUByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Skipping test because testDB is not initialized")
	}
	tx := testDB.Begin()
	defer tx.Rollback()
	repo := repository.NewProductRepository(tx)
	ctx := context.Background()

	spu, err := createRandomSPU(ctx, repo)
	require.NoError(t, err)
	require.NotEmpty(t, spu.SKUs, "Should have created at least one SKU")
	sku1 := spu.SKUs[0]

	t.Run("Success", func(t *testing.T) {
		sku2, err := repo.GetSKUByID(ctx, sku1.ID)
		require.NoError(t, err)
		require.NotNil(t, sku2)

		assert.Equal(t, sku1.ID, sku2.ID)
		assert.Equal(t, sku1.SPUID, sku2.SPUID)
		assert.Equal(t, sku1.Attributes, sku2.Attributes)
	})

	t.Run("NotFound", func(t *testing.T) {
		sku3, err := repo.GetSKUByID(ctx, nonExistentID)
		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrSKUNotFound) // Assert sentinel error
		require.Nil(t, sku3)
	})
}