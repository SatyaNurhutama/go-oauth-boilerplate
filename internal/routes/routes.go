package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/handler"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/middleware"
)

func SetupRoutes(router *gin.Engine, authHandler *handler.AuthHandler, redisClient *redis.Client) {
	// Public routes (no authentication required)
	public := router.Group("/api")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
		public.GET("/auth/login/google", authHandler.GoogleLogin)
		public.GET("/auth/login/google/callback", authHandler.GoogleCallback)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(redisClient))
	{
		protected.POST("/auth/logout", authHandler.Logout)
		protected.POST("/auth/refresh", authHandler.RefreshToken)
	}
}
