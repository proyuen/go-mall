package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/service"
	"github.com/proyuen/go-mall/pkg/utils"
)

// OrderHandler defines the HTTP handlers for order-related operations.
type OrderHandler struct {
	orderService service.OrderService
}

// NewOrderHandler creates a new OrderHandler instance.
func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrderRequest defines the request body for creating an order.
type CreateOrderRequest struct {
	Items []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateOrderItemRequest defines the request body for an item within an order.
type CreateOrderItemRequest struct {
	SKUID    uint64 `json:"sku_id" binding:"required,gt=0"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
}

// CreateOrder handles the creation of a new order.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": err.Error()})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": err.Error()})
		return
	}

	// Map handler request DTO to service request DTO
	var serviceItems []service.OrderItemReq
	for _, item := range req.Items {
		serviceItems = append(serviceItems, service.OrderItemReq{
			SKUID:    item.SKUID,
			Quantity: item.Quantity,
		})
	}

	serviceReq := &service.OrderCreateReq{
		UserID: userID,
		Items:  serviceItems,
	}

	resp, err := h.orderService.CreateOrder(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": http.StatusCreated, "message": "Order created successfully", "data": resp})
}