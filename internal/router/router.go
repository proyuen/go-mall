package router

import (
	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/handler"
	"github.com/proyuen/go-mall/internal/middleware"
)

// Router struct holds dependencies for routing.
type Router struct {
	userHandler    *handler.UserHandler
	productHandler *handler.ProductHandler
	jwtSecret      string
}

// NewRouter creates a new Router instance.
func NewRouter(userHandler *handler.UserHandler, productHandler *handler.ProductHandler, jwtSecret string) *Router {
	return &Router{
		userHandler:    userHandler,
		productHandler: productHandler,
		jwtSecret:      jwtSecret,
	}
}

// InitRoutes initializes all application routes.
func (r *Router) InitRoutes() *gin.Engine {
	engine := gin.Default()

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
			// Use AuthMiddleware to protect product creation
			productRoutes.POST("", middleware.AuthMiddleware(r.jwtSecret), r.productHandler.CreateProduct)
			
			// Public routes
			productRoutes.GET("/:id", r.productHandler.GetProduct)
			productRoutes.GET("", r.productHandler.ListProducts)
		}
	}

	return engine
}
