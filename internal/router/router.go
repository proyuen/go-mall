package router

import (
	"github.com/gin-gonic/gin"
	"github.com/proyuen/go-mall/internal/handler"
)

// Router struct holds dependencies for routing.
type Router struct {
	userHandler *handler.UserHandler
}

// NewRouter creates a new Router instance.
func NewRouter(userHandler *handler.UserHandler) *Router {
	return &Router{
		userHandler: userHandler,
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
	}

	return engine
}
