package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/proyuen/go-mall/internal/handler"
	"github.com/proyuen/go-mall/internal/middleware"
	"github.com/proyuen/go-mall/pkg/token"
)

// Router struct holds dependencies for routing.
type Router struct {
	userHandler    *handler.UserHandler
	productHandler *handler.ProductHandler
	orderHandler   *handler.OrderHandler
	tokenMaker     token.Maker
}

// NewRouter creates a new Router instance.
func NewRouter(userHandler *handler.UserHandler, productHandler *handler.ProductHandler, orderHandler *handler.OrderHandler, tokenMaker token.Maker) *Router {
	return &Router{
		userHandler:    userHandler,
		productHandler: productHandler,
		orderHandler:   orderHandler,
		tokenMaker:     tokenMaker,
	}
}

// InitRoutes initializes all application routes.
func (r *Router) InitRoutes() *gin.Engine {
	engine := gin.Default()

	// Metrics endpoint
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API Group for version 1
	v1 := engine.Group("/api/v1")
	{
		// User routes
		userRoutes := v1.Group("/users")
		{
			userRoutes.POST("/register", r.userHandler.Register)
			userRoutes.POST("/login", r.userHandler.Login)
		}

		// Product routes
		productRoutes := v1.Group("/products")
		{
			// Protected routes
			productRoutes.POST("", middleware.AuthMiddleware(r.tokenMaker), r.productHandler.CreateProduct)
			
			// Public routes
			productRoutes.GET("/:id", r.productHandler.GetProduct)
			productRoutes.GET("", r.productHandler.ListProducts)
		}

		// Order routes (All protected)
		orderRoutes := v1.Group("/orders")
		orderRoutes.Use(middleware.AuthMiddleware(r.tokenMaker))
		{
			orderRoutes.POST("", r.orderHandler.CreateOrder)
		}
	}

	return engine
}