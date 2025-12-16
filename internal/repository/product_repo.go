package repository

import (
	"context"
	"errors" // Import errors package
	"fmt"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/pkg/database"
	"gorm.io/gorm"
)

// ErrSPUNotFound is returned when an SPU record is not found.
var ErrSPUNotFound = errors.New("SPU not found")

// ErrSKUNotFound is returned when an SKU record is not found.
var ErrSKUNotFound = errors.New("SKU not found")

//go:generate mockgen -source=$GOFILE -destination=../mocks/product_repo_mock.go -package=mocks
// ProductRepository defines the interface for product data operations.
type ProductRepository interface {
	CreateSPU(ctx context.Context, spu *model.SPU) error
	CreateSKU(ctx context.Context, sku *model.SKU) error
	GetSPUByID(ctx context.Context, id uint64) (*model.SPU, error)
	GetSKUByID(ctx context.Context, id uint64) (*model.SKU, error)
	ListSPUs(ctx context.Context, offset, limit int) ([]model.SPU, error)
	UpdateSKUStock(ctx context.Context, skuID uint64, quantity int) error
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
	db := database.GetDBFromContext(ctx, r.db)
	if err := db.Create(spu).Error; err != nil {
		return fmt.Errorf("failed to create SPU: %w", err)
	}
	return nil
}

// CreateSKU saves a new SKU to the database.
func (r *productRepository) CreateSKU(ctx context.Context, sku *model.SKU) error {
	db := database.GetDBFromContext(ctx, r.db)
	if err := db.Create(sku).Error; err != nil {
		return fmt.Errorf("failed to create SKU: %w", err)
	}
	return nil
}

// GetSPUByID retrieves an SPU by its ID.
func (r *productRepository) GetSPUByID(ctx context.Context, id uint64) (*model.SPU, error) {
	var spu model.SPU
	db := database.GetDBFromContext(ctx, r.db)
	if err := db.First(&spu, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSPUNotFound
		}
		return nil, fmt.Errorf("failed to get SPU by ID '%d': %w", id, err)
	}
	return &spu, nil
}

// GetSKUByID retrieves an SKU by its ID. It also preloads the associated SPU.
func (r *productRepository) GetSKUByID(ctx context.Context, id uint64) (*model.SKU, error) {
	var sku model.SKU
	db := database.GetDBFromContext(ctx, r.db)
	// Preload SPU to get product name/description directly
	if err := db.Preload("SPU").First(&sku, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSKUNotFound
		}
		return nil, fmt.Errorf("failed to get SKU by ID '%d': %w", id, err)
	}
	return &sku, nil
}

// ListSPUs retrieves a list of SPUs with pagination.
func (r *productRepository) ListSPUs(ctx context.Context, offset, limit int) ([]model.SPU, error) {
	// Limit cap protection to prevent OOM
	if limit > 100 {
		limit = 100
	}
	
	var spuList []model.SPU
	db := database.GetDBFromContext(ctx, r.db)
	// Deterministic ordering to prevent random results
	if err := db.Order("id DESC").Offset(offset).Limit(limit).Find(&spuList).Error; err != nil {
		return nil, fmt.Errorf("failed to list SPUs: %w", err)
	}
	return spuList, nil
}

// UpdateSKUStock deducts/adds stock for a given SKU.
// quantity can be negative for deduction, positive for addition.
// It ensures stock does not go below zero.
func (r *productRepository) UpdateSKUStock(ctx context.Context, skuID uint64, quantity int) error {
	db := database.GetDBFromContext(ctx, r.db)
	result := db.Model(&model.SKU{}).
		Where("id = ? AND stock >= ?", skuID, -quantity). // Ensure sufficient stock for deduction
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity))

	if result.Error != nil {
		return fmt.Errorf("failed to update SKU stock for ID '%d': %w", skuID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("not enough stock or SKU ID '%d' not found", skuID)
	}
	return nil
}
