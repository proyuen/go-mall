package service

import (
	"context"
	"fmt"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/repository"
)

// ProductCreateReq defines the request structure for creating a new product.
type ProductCreateReq struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CategoryID  uint              `json:"category_id"`
	SKUs        []SKUCreateReq    `json:"skus"` // List of SKUs for this product
}

// SKUCreateReq defines the request structure for creating an SKU within a product.
type SKUCreateReq struct {
	Attributes string  `json:"attributes"` // JSON string for attributes like {"color": "red", "size": "M"}
	Price      float64 `json:"price"`
	Stock      int     `json:"stock"`
	Image      string  `json:"image"`
}

// ProductCreateResp defines the response structure after creating a product.
type ProductCreateResp struct {
	SPUID uint `json:"spu_id"`
}

// ProductResp defines the response structure for a product (SPU with its SKUs).
type ProductResp struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	CategoryID  uint         `json:"category_id"`
	SKUs        []SKUResp    `json:"skus"`
}

// SKUResp defines the response structure for an SKU.
type SKUResp struct {
	ID         uint    `json:"id"`
	Attributes string  `json:"attributes"`
	Price      float64 `json:"price"`
	Stock      int     `json:"stock"`
	Image      string  `json:"image"`
}

// ProductService defines the interface for product business logic.
type ProductService interface {
	CreateProduct(ctx context.Context, req *ProductCreateReq) (*ProductCreateResp, error)
	GetProduct(ctx context.Context, spuID uint) (*ProductResp, error)
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
		skus = append(skus, model.SKU{
			Attributes: skuReq.Attributes,
			Price:      skuReq.Price,
			Stock:      skuReq.Stock,
			Image:      skuReq.Image,
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
func (s *productService) GetProduct(ctx context.Context, spuID uint) (*ProductResp, error) {
	spu, err := s.repo.GetSPUByID(ctx, spuID)
	if err != nil {
		return nil, err
	}

	// TODO: Fetch SKUs for this SPU. Current GetSKUByID is for single SKU.
	return &ProductResp{
		ID:          spu.ID,
		Name:        spu.Name,
		Description: spu.Description,
		CategoryID:  spu.CategoryID,
		SKUs:        []SKUResp{}, // Placeholder
	}, nil
}

// ListProducts retrieves a list of products (SPUs) with pagination.
func (s *productService) ListProducts(ctx context.Context, offset, limit int) ([]ProductResp, error) {
	spuList, err := s.repo.ListSPUs(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	var productResps []ProductResp
	for _, spu := range spuList {
		productResps = append(productResps, ProductResp{
			ID:          spu.ID,
			Name:        spu.Name,
			Description: spu.Description,
			CategoryID:  spu.CategoryID,
			SKUs:        []SKUResp{}, // Placeholder
		})
	}
	return productResps, nil
}