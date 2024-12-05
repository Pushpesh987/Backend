package router

import (
	"fmt"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	// "Backend/src/middleware"         // Custom middleware
	"Backend/src/core/middleware" // Authentication module
	"Backend/src/modules/authentication"
	"Backend/src/modules/events"
	"Backend/src/modules/feed"
	"Backend/src/modules/posts"
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
	postGroup := router.Group("/posts")
	feedGroup := router.Group("/feed")
	eventGroup:= router.Group("/events")
	// messagesGroup := router.Group("/messages")

	// Authentication routes
	authGroup.Post("/signup", authentication.SignUp)
	authGroup.Post("/signin", authentication.SignIn)
	// authGroup.Post("/reset-password", middleware.Protected(), auth.ResetPassword)

	// User routes
	userGroup.Get("/profile", middleware.Protected(), users.GetProfile)
	userGroup.Put("/Update-profile", middleware.Protected(), users.UpdateProfile)
    userGroup.Post("/profile", middleware.Protected(), users.CreateProfile)
	userGroup.Post("/upload-profile-photo", middleware.Protected(), users.UploadProfilePhoto)
	userGroup.Post("/update-skill-interest", middleware.Protected(), users.UpdateUserSkillsAndInterests)


	postGroup.Post("/post",middleware.Protected(),posts.CreatePost)
    postGroup.Post("/like",middleware.Protected(),posts.CreateLike)
	postGroup.Post("/comment",middleware.Protected(),posts.CreateComment)
	postGroup.Get("/:post_id/likes/count",middleware.Protected(),posts.GetLikesCount)
	postGroup.Post("/share",middleware.Protected(),posts.CreateShare)

	eventGroup.Post("/event",middleware.Protected(),events.CreateEvent)
	eventGroup.Post("/workshop",middleware.Protected(),events.CreateWorkshop)
	eventGroup.Post("/project",middleware.Protected(),events.CreateProject)
	eventGroup.Get("/event/:id",middleware.Protected(),events.GetEventByID)
	eventGroup.Get("/workshop/:id",middleware.Protected(),events.GetWorkshopByID)
	eventGroup.Get("/project/:id",middleware.Protected(),events.GetProjectByID)
	eventGroup.Get("/eventsfeed",middleware.Protected(),events.GetEventsFeed)
	eventGroup.Get("/workshopsfeed",middleware.Protected(),events.GetWorkshopsFeed)
	eventGroup.Get("/projectsfeed",middleware.Protected(),events.GetProjectsFeed)


	// // Feed routes
	feedGroup.Get("/", middleware.Protected(), feed.FetchFeed)
	// feedGroup.Post("/", middleware.Protected(), feed.CreatePost)
	// feedGroup.Delete("/:post_id", middleware.Protected(), feed.DeletePost)

	// // Messaging routes
	// messagesGroup.Get("/:user_id", middleware.Protected(), messages.GetMessages)
	// messagesGroup.Post("/:user_id", middleware.Protected(), messages.SendMessage)
}
