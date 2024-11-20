package router

import (
	"fmt"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	// "Backend/src/middleware"         // Custom middleware
	"Backend/src/core/middleware" // Authentication module
	"Backend/src/modules/authentication"
	"Backend/src/modules/users" // User module
	// "Backend/src/modules/authentication"
	// "your_project_name/src/modules/feed"       // Feed module
	// "your_project_name/src/modules/messages"   // Messaging module
)

func InitialiseAndSetupRoutes(app *fiber.App) {
	// Root logger middleware for monitoring requests
	root := app.Group("/", logger.New())

	// Simple ping route for health checks
	root.Get("/ping", func(c *fiber.Ctx) error { return c.SendString("pong") })

	// API version grouping
	apiV1 := root.Group("/api/v1")
	setupAPIV1Routes(apiV1)

	// Catch-all route for debugging load balancer traffic
	root.Get("/:any", func(c *fiber.Ctx) error {
		return c.SendString(c.Params("any"))
	})

	// Display all registered routes for debugging
	routes := app.GetRoutes()
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})
	for _, route := range routes {
		fmt.Printf("%s\t%s\n", route.Method, route.Path)
	}
}

func setupAPIV1Routes(router fiber.Router) {
	// Grouped API endpoints
	authGroup := router.Group("/auth")
	userGroup := router.Group("/users")
	// feedGroup := router.Group("/feed")
	// messagesGroup := router.Group("/messages")

	// Authentication routes
	authGroup.Post("/signup", authentication.SignUp)
	authGroup.Post("/signin", authentication.SignIn)
	// authGroup.Post("/reset-password", middleware.Protected(), auth.ResetPassword)

	// User routes
	userGroup.Get("/:id/profile", middleware.Protected(), users.GetUserDetails)
	userGroup.Put("/:id/profile", middleware.Protected(), users.UpdateUserDetails)

	// // Feed routes
	// feedGroup.Get("/", middleware.Protected(), feed.GetFeed)
	// feedGroup.Post("/", middleware.Protected(), feed.CreatePost)
	// feedGroup.Delete("/:post_id", middleware.Protected(), feed.DeletePost)

	// // Messaging routes
	// messagesGroup.Get("/:user_id", middleware.Protected(), messages.GetMessages)
	// messagesGroup.Post("/:user_id", middleware.Protected(), messages.SendMessage)
}
