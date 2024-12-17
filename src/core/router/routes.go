package router

import (
	"Backend/src/core/middleware"
	"Backend/src/modules/IoT_logs"
	"Backend/src/modules/authentication"
	"Backend/src/modules/communities"
	connection "Backend/src/modules/connections"
	"Backend/src/modules/events"
	"Backend/src/modules/feed"
	"Backend/src/modules/messages"
	"Backend/src/modules/notifications"
	"Backend/src/modules/posts"
	"Backend/src/modules/questions"
	// "Backend/src/modules/communities"
	"Backend/src/modules/users"
	"fmt"
	"sort"
	"github.com/gofiber/websocket/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func InitialiseAndSetupRoutes(app *fiber.App) {
	root := app.Group("/", logger.New())

	root.Get("/ping", func(c *fiber.Ctx) error { return c.SendString("pong") })

	apiV1 := root.Group("/api/v1")
	setupAPIV1Routes(apiV1)

	root.Get("/:any", func(c *fiber.Ctx) error {
		return c.SendString(c.Params("any"))
	})

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
	eventGroup := router.Group("/events")
	questionGroup := router.Group("/question")
	communityGroup :=router.Group("/communities")
	iotlogsGroup :=router.Group("/iotlogs")
	notificationsGroup :=router.Group("/notification")
	// messagesGroup := router.Group("/messages")

	// Authentication routes
	authGroup.Post("/signup", authentication.SignUp)
	authGroup.Post("/signin", authentication.SignIn)
	// authGroup.Post("/reset-password", middleware.Protected(), auth.ResetPassword)

	// User routes
	userGroup.Get("/profile", middleware.Protected(), users.GetProfile)
	userGroup.Put("/Update-profile", middleware.Protected(), users.UpdateProfile)
	// userGroup.Post("/profile", middleware.Protected(), users.CreateProfile)
	userGroup.Post("/upload-profile-photo", middleware.Protected(), users.UploadProfilePhoto)
	// userGroup.Post("/update-skill-interest", middleware.Protected(), users.UpdateUserSkillsAndInterests)
	userGroup.Post("/follow", middleware.Protected(), connection.Follow)
	userGroup.Post("/check-connection", middleware.Protected(), connection.ConnectionCheck)
	userGroup.Get("/location", middleware.Protected(), users.GetAllLocationNames)
	userGroup.Get("/skills", middleware.Protected(), users.GetAllSkills)
	userGroup.Get("/interests", middleware.Protected(), users.GetAllInterests)
	userGroup.Get("/education-level", middleware.Protected(), users.GetAllFieldsOfStudy)
	userGroup.Get("/fields-of-study", middleware.Protected(), users.GetAllEducationLevels)
	userGroup.Get("/college", middleware.Protected(), users.GetAllColleges)
	userGroup.Get("/search",users.SearchUsers)
	userGroup.Get("/profile/:id",middleware.Protected(),users.GetProfileByID)
	
	postGroup.Post("/post", middleware.Protected(), posts.CreatePost)
	postGroup.Post("/like", middleware.Protected(), posts.CreateLike)
	postGroup.Post("/comment", middleware.Protected(), posts.CreateComment)
	postGroup.Get("/:post_id/likes/count", middleware.Protected(), posts.GetLikesCount)
	postGroup.Post("/share", middleware.Protected(), posts.CreateShare)

	eventGroup.Post("/event", middleware.Protected(), events.CreateEvent)
	eventGroup.Post("/workshop", middleware.Protected(), events.CreateWorkshop)
	eventGroup.Post("/project", middleware.Protected(), events.CreateProject)
	eventGroup.Get("/event/:id", middleware.Protected(), events.GetEventByID)
	eventGroup.Get("/workshop/:id", middleware.Protected(), events.GetWorkshopByID)
	eventGroup.Get("/project/:id", middleware.Protected(), events.GetProjectByID)
	eventGroup.Get("/eventsfeed", middleware.Protected(), events.GetEventsFeed)
	eventGroup.Get("/workshopsfeed", middleware.Protected(), events.GetWorkshopsFeed)
	eventGroup.Get("/projectsfeed", middleware.Protected(), events.GetProjectsFeed)

	questionGroup.Get("/daily", middleware.Protected(), questions.GetDailyQuestions)
	questionGroup.Get("/skill", middleware.Protected(), questions.GetSkillQuestions)
	questionGroup.Get("/bonus", middleware.Protected(), questions.GetBonusQuestions)
	questionGroup.Post("/submit", middleware.Protected(), questions.SubmitAnswer)

	communityGroup.Post("/create",middleware.Protected(),communities.CreateCommunity)
	communityGroup.Post("/:id/join",middleware.Protected(),communities.JoinCommunity)
	communityGroup.Get("/:id",middleware.Protected(), communities.GetCommunityDetails)
    communityGroup.Get("/", middleware.Protected(), communities.GetAllCommunities)
	communityGroup.Post("/:id/leave", middleware.Protected(), communities.LeaveCommunity)
	communityGroup.Get("/user-joined", middleware.Protected(), communities.GetUserCommunities)
	communityGroup.Get("/:id/messages", middleware.Protected(), communities.GetCommunityMessages)
	communityGroup.Post("/:id/messages", middleware.Protected(), messages.SendMessage)
	communityGroup.Get("/:id/messages/ws", middleware.Protected(), messages.WebSocketHandler)

	iotlogsGroup.Post("/",iotlogs.CreateIotLog)

	notificationsGroup.Get("/ws",websocket.New(notifications.NotificationWebSocketHandler))

	// // Feed routes
	feedGroup.Get("/", middleware.Protected(), feed.FetchFeed)
	// feedGroup.Post("/", middleware.Protected(), feed.CreatePost)
	// feedGroup.Delete("/:post_id", middleware.Protected(), feed.DeletePost)

	// // Messaging routes
	// messagesGroup.Get("/:user_id", middleware.Protected(), messages.GetMessages)
	// messagesGroup.Post("/:user_id", middleware.Protected(), messages.SendMessage)
}
