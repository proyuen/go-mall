package router

import (
	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/handler"
)

// Router struct holds dependencies for routing.
type Router struct {
	userHandler    *handler.UserHandler
	productHandler *handler.ProductHandler
}

// NewRouter creates a new Router instance.
func NewRouter(userHandler *handler.UserHandler, productHandler *handler.ProductHandler) *Router {
	return &Router{
		userHandler:    userHandler,
		productHandler: productHandler,
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
			// TODO: Protect CreateProduct with Auth Middleware
			productRoutes.POST("", r.productHandler.CreateProduct)
			productRoutes.GET("/:id", r.productHandler.GetProduct)
			productRoutes.GET("", r.productHandler.ListProducts)
		}
	}

	return engine
}