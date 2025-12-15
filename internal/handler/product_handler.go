package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/service"
)

// ProductHandler defines the HTTP handlers for product-related operations.
type ProductHandler struct {
	productService service.ProductService
}

// NewProductHandler creates a new ProductHandler instance.
func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// CreateProductRequest defines the request body for creating a product.
type CreateProductRequest struct {
	Name        string       `json:"name" binding:"required"`
	Description string       `json:"description"`
	CategoryID  uint         `json:"category_id" binding:"required"`
	SKUs        []SKURequest `json:"skus" binding:"required,dive"` // dive validates items in the slice
}

type SKURequest struct {
	Attributes string  `json:"attributes" binding:"required"`
	Price      float64 `json:"price" binding:"required,gt=0"`
	Stock      int     `json:"stock" binding:"required,gte=0"`
	Image      string  `json:"image"`
}

// CreateProduct handles the creation of a new product.
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": err.Error()})
		return
	}

	// Map handler request DTO to service request DTO
	var skus []service.SKUCreateReq
	for _, sku := range req.SKUs {
		skus = append(skus, service.SKUCreateReq{
			Attributes: sku.Attributes,
			Price:      sku.Price,
			Stock:      sku.Stock,
			Image:      sku.Image,
		})
	}

	serviceReq := &service.ProductCreateReq{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		SKUs:        skus,
	}

	resp, err := h.productService.CreateProduct(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": http.StatusCreated, "message": "Product created successfully", "data": resp})
}

// GetProduct retrieves a product by its ID.
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "invalid product id"})
		return
	}

	resp, err := h.productService.GetProduct(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "success", "data": resp})
}

// ListProducts retrieves a list of products with pagination.
func (h *ProductHandler) ListProducts(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "invalid offset"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "invalid limit"})
		return
	}

	resp, err := h.productService.ListProducts(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "success", "data": resp})
}
