package main

import (
	"goapi/internal/config"
	"goapi/internal/handlers"
	"goapi/internal/middleware"
	"goapi/internal/models"
	"goapi/internal/repository"
	"goapi/internal/services"
	"log"
	"time"

	"fmt"

	"goapi/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Logger
	logger.Init()

	// Load config
	cfg := config.Load()

	// Initialize database
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Redis
	redisClient, err := config.InitRedis(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Auto-migrate models
	log.Println("Run database migration...")
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repository, service, handler
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo, redisClient)
	userHandler := handlers.NewUserHandler(userService)

	// Setup Gin router (Use New() to avoid default Logger)
	router := gin.New()
	router.Use(gin.Recovery())

	// Global middleware
	router.Use(middleware.RequestID()) // Add Request ID first
	router.Use(middleware.Logger())    // Add Custom Logger
	router.Use(middleware.CORS())

	// Global Rate Limiter: 100 requests per minute
	router.Use(middleware.RateLimiter(redisClient, 100, time.Minute))

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// API routes v1
	v1 := router.Group("/api/v1")
	{
		// Public routes
		// Strict Rate Limiter for Auth: 5 requests per minute
		authLimiter := middleware.RateLimiter(redisClient, 5, time.Minute)

		v1.POST("/register", authLimiter, userHandler.Register)
		v1.POST("/login", authLimiter, userHandler.Login)

		// Protected routes
		authorized := v1.Group("")
		authorized.Use(middleware.JWTAuth())
		{
			authorized.GET("/users", userHandler.GetAllUsers)
			authorized.GET("/users/:id", userHandler.GetUserByID)
			authorized.PUT("/users/:id", userHandler.UpdateUser)
			authorized.DELETE("/users/:id", userHandler.DeleteUser)
			authorized.GET("/me", userHandler.GetCurrentUser)
		}
	}

	fmt.Println(`
 ______     ______        ______     ______   __    
/\  ___\   /\  __ \      /\  __ \   /\  == \ /\ \   
\ \ \__ \  \ \ \/\ \     \ \  __ \  \ \  _-/ \ \ \  
 \ \_____\  \ \_____\     \ \_\ \_\  \ \_\    \ \_\ 
  \/_____/   \/_____/      \/_/\/_/   \/_/     \/_/ `)

	// Run server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
