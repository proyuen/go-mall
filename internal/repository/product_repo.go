package repository

import (
	"context"
	"fmt"

	"github.com/proyuen/go-mall/internal/model"
	"gorm.io/gorm"
)

// ProductRepository defines the interface for product data operations.
type ProductRepository interface {
	CreateSPU(ctx context.Context, spu *model.SPU) error
	CreateSKU(ctx context.Context, sku *model.SKU) error
	GetSPUByID(ctx context.Context, id uint) (*model.SPU, error)
	GetSKUByID(ctx context.Context, id uint) (*model.SKU, error)
	ListSPUs(ctx context.Context, offset, limit int) ([]model.SPU, error)
}

// productRepository implements ProductRepository using GORM.
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository instance.
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

// CreateSPU saves a new SPU to the database.
func (r *productRepository) CreateSPU(ctx context.Context, spu *model.SPU) error {
	if err := r.db.WithContext(ctx).Create(spu).Error; err != nil {
		return fmt.Errorf("failed to create SPU: %w", err)
	}
	return nil
}

// CreateSKU saves a new SKU to the database.
func (r *productRepository) CreateSKU(ctx context.Context, sku *model.SKU) error {
	if err := r.db.WithContext(ctx).Create(sku).Error; err != nil {
		return fmt.Errorf("failed to create SKU: %w", err)
	}
	return nil
}

// GetSPUByID retrieves an SPU by its ID.
func (r *productRepository) GetSPUByID(ctx context.Context, id uint) (*model.SPU, error) {
	var spu model.SPU
	if err := r.db.WithContext(ctx).First(&spu, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("SPU with ID '%d' not found", id)
		}
		return nil, fmt.Errorf("failed to get SPU by ID '%d': %w", id, err)
	}
	return &spu, nil
}

// GetSKUByID retrieves an SKU by its ID. It also preloads the associated SPU.
func (r *productRepository) GetSKUByID(ctx context.Context, id uint) (*model.SKU, error) {
	var sku model.SKU
	// Preload SPU to get product name/description directly
	if err := r.db.WithContext(ctx).Preload("SPU").First(&sku, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("SKU with ID '%d' not found", id)
		}
		return nil, fmt.Errorf("failed to get SKU by ID '%d': %w", id, err)
	}
	return &sku, nil
}

// ListSPUs retrieves a list of SPUs with pagination.
func (r *productRepository) ListSPUs(ctx context.Context, offset, limit int) ([]model.SPU, error) {
	var spuList []model.SPU
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&spuList).Error; err != nil {
		return nil, fmt.Errorf("failed to list SPUs: %w", err)
	}
	return spuList, nil
}
