package main

import (
	"goapi/internal/config"
	"goapi/internal/handlers"
	"goapi/internal/middleware"
	"goapi/internal/models"
	"goapi/internal/repository"
	"goapi/internal/services"
	"log"

	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg := config.Load()

	// Initialize database
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	log.Println("Run database migration...")
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repository, service, handler
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	// Setup Gin router
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// API routes v1
	v1 := router.Group("/api/v1")
	{
		// Public routes
		v1.POST("/register", userHandler.Register)
		v1.POST("/login", userHandler.Login)

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
