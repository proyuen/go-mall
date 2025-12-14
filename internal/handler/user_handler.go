package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/service"
)

// UserHandler defines the HTTP handlers for user-related operations.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// RegisterRequest defines the request body for user registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// Register handles user registration.
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": err.Error()})
		return
	}

	// Map handler request DTO to service request DTO
	serviceReq := &service.UserRegisterReq{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.userService.Register(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "User registered successfully", "data": resp})
}

// LoginRequest defines the request body for user login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login and returns a JWT token.
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": err.Error()})
		return
	}

	// Map handler request DTO to service request DTO
	serviceReq := &service.UserLoginReq{
		Username: req.Username,
		Password: req.Password,
	}

	resp, err := h.userService.Login(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "Login successful", "data": resp})
}
