package service

import (
	"context"
	"encoding/json"
	"errors" // Added errors import
	"fmt"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm" // Added gorm import
)

// ProductCreateReq defines the request structure for creating a new product.
type ProductCreateReq struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CategoryID  uint64            `json:"category_id,string"` // Changed to uint64
	SKUs        []SKUCreateReq    `json:"skus"`               // List of SKUs for this product
}

// SKUCreateReq defines the request structure for creating an SKU within a product.
type SKUCreateReq struct {
	Attributes json.RawMessage `json:"attributes"` // Use RawMessage for flexibility, will unmarshal to model.JSONB
	Price      decimal.Decimal `json:"price"`      // Changed to decimal.Decimal
	Stock      int             `json:"stock"`
	// Image removed as per model definition
}

// ProductCreateResp defines the response structure after creating a product.
type ProductCreateResp struct {
	SPUID uint64 `json:"spu_id,string"` // Changed to uint64
}

// ProductResp defines the response structure for a product (SPU with its SKUs).
type ProductResp struct {
	ID          uint64        `json:"id,string"`          // Changed to uint64
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CategoryID  uint64        `json:"category_id,string"` // Changed to uint64
	SKUs        []SKUResp     `json:"skus"`
}

// SKUResp defines the response structure for an SKU.
type SKUResp struct {
	ID         uint64          `json:"id,string"`      // Changed to uint64
	Attributes model.JSONB     `json:"attributes"` // Changed to model.JSONB for response
	Price      decimal.Decimal `json:"price"`      // Changed to decimal.Decimal
	Stock      int             `json:"stock"`
	// Image removed as per model definition
}

//go:generate mockgen -source=$GOFILE -destination=../mocks/product_service_mock.go -package=mocks
// ProductService defines the interface for product business logic.
type ProductService interface {
	CreateProduct(ctx context.Context, req *ProductCreateReq) (*ProductCreateResp, error)
	GetProduct(ctx context.Context, spuID uint64) (*ProductResp, error) // Changed to uint64
	ListProducts(ctx context.Context, offset, limit int) ([]ProductResp, error)
}

type productService struct {
	repo repository.ProductRepository
}

// NewProductService creates a new ProductService instance.
func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

// CreateProduct creates a new SPU and its associated SKUs in a single transaction.
func (s *productService) CreateProduct(ctx context.Context, req *ProductCreateReq) (*ProductCreateResp, error) {
	// Assemble SKUs
	var skus []model.SKU
	for _, skuReq := range req.SKUs {
		var attributes model.JSONB
		if skuReq.Attributes != nil { // Check if attributes are provided
			if err := json.Unmarshal(skuReq.Attributes, &attributes); err != nil {
				return nil, fmt.Errorf("invalid SKU attributes JSON: %w", err)
			}
		}

		skus = append(skus, model.SKU{
			Attributes: attributes,
			Price:      skuReq.Price,
			Stock:      skuReq.Stock,
			// Image removed
		})
	}

	// Assemble SPU with embedded SKUs
	spu := &model.SPU{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		SKUs:        skus, // GORM will handle the association creation
	}

	// Save SPU (and SKUs automatically via GORM association)
	if err := s.repo.CreateSPU(ctx, spu); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &ProductCreateResp{SPUID: spu.ID}, nil
}

// GetProduct retrieves a product (SPU) with all its associated SKUs.
func (s *productService) GetProduct(ctx context.Context, spuID uint64) (*ProductResp, error) { // Changed to uint64
	spu, err := s.repo.GetSPUByID(ctx, spuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // Assuming repo returns gorm.ErrRecordNotFound
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get SPU by ID %d: %w", spuID, err)
	}

	var skuResps []SKUResp
	// Assuming spu.SKUs is preloaded by the repository (or we fetch them explicitly)
	// For now, let's assume the repo will preload them based on the model definition.
	for _, sku := range spu.SKUs {
		skuResps = append(skuResps, SKUResp{
			ID:         sku.ID,
			Attributes: sku.Attributes,
			Price:      sku.Price,
			Stock:      sku.Stock,
		})
	}

	return &ProductResp{
		ID:          spu.ID,
		Name:        spu.Name,
		Description: spu.Description,
		CategoryID:  spu.CategoryID,
		SKUs:        skuResps,
	}, nil
}

// ListProducts retrieves a list of products (SPUs) with pagination.
func (s *productService) ListProducts(ctx context.Context, offset, limit int) ([]ProductResp, error) {
	spuList, err := s.repo.ListSPUs(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list SPUs: %w", err)
	}

	var productResps []ProductResp
	for _, spu := range spuList {
		var skuResps []SKUResp
		for _, sku := range spu.SKUs { // Assume SKUs are preloaded
			skuResps = append(skuResps, SKUResp{
				ID:         sku.ID,
				Attributes: sku.Attributes,
				Price:      sku.Price,
				Stock:      sku.Stock,
			})
		}

		productResps = append(productResps, ProductResp{
			ID:          spu.ID,
			Name:        spu.Name,
			Description: spu.Description,
			CategoryID:  spu.CategoryID,
			SKUs:        skuResps,
		})
	}
	return productResps, nil
}
