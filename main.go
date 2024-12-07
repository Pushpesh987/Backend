package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"Backend/src/core/config"
    "Backend/src/core/database"
    "Backend/src/core/router"
)

func main() {
	// Initialize the Fiber app
	app := fiber.New()

	// Middleware
	app.Use(recover.New())      // Recover middleware to handle panics
	app.Use(cors.New())         // CORS middleware for cross-origin requests
	app.Use(requestid.New())    // Middleware to generate unique request IDs

	// Setup environment variables
	config.SetupEnv()

	// Connect to the database
	database.ConnectDB()

	// Set up routes
	router.InitialiseAndSetupRoutes(app)

	// Get port from environment variable, default to 3000
	port := config.Config("PORT") // Render provides this
		if port == "" {
			port = "3000" // Default fallback
		}

	// port := config.Config("PORT") // Render provides this
	// if port == "" {
	// 	port = config.Config("APP_PORT") // Use APP_PORT if locally testing
	// 	if port == "" {
	// 		port = "3000" // Default fallback
	// 	}
	// }

	log.Printf("Starting server on port %s...", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}
