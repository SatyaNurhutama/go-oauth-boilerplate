package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/handler"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/repository"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/usecase"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/config"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/routes"
	"github.com/satya-nurhutama/go-oauth-boilerplate/pkg/database"
)

func main() {

	config.LoadEnv()

	// Initialize the database connection
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize the Redis client
	redisClient, err := database.NewRedisClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Define module
	userRepo := repository.NewUserRepository(db)
	authUseCase := usecase.NewAuthUseCase(*userRepo, redisClient)
	authHandler := handler.NewAuthHandler(*authUseCase)

	router := gin.Default()
	routes.SetupRoutes(router, authHandler, redisClient)

	// Start the server
	log.Printf("Server started on :%s", config.GetEnv("PORT"))
	if err := http.ListenAndServe(":"+config.GetEnv("PORT"), router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
