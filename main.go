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

	// Get port from config and start the server
	port := config.Config("APP_PORT")
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}
