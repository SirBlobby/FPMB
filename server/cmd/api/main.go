package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fpmb/server/internal/database"
	"github.com/fpmb/server/internal/handlers"
	"github.com/fpmb/server/internal/middleware"
	"github.com/fpmb/server/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func startDueDateReminder() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		runDueDateReminder()
		for range ticker.C {
			runDueDateReminder()
		}
	}()
}

func runDueDateReminder() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	thresholds := []int{1, 3}

	for _, days := range thresholds {
		windowStart := now.Add(time.Duration(days)*24*time.Hour - 30*time.Minute)
		windowEnd := now.Add(time.Duration(days)*24*time.Hour + 30*time.Minute)

		cursor, err := database.GetCollection("cards").Find(ctx, bson.M{
			"due_date": bson.M{"$gte": windowStart, "$lte": windowEnd},
		})
		if err != nil {
			continue
		}

		var cards []models.Card
		cursor.All(ctx, &cards)
		cursor.Close(ctx)

		for _, card := range cards {
			for _, email := range card.Assignees {
				var user models.User
				if err := database.GetCollection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
					continue
				}

				cutoff := now.Add(-24 * time.Hour)
				count, _ := database.GetCollection("notifications").CountDocuments(ctx, bson.M{
					"user_id":    user.ID,
					"type":       "due_soon",
					"card_id":    card.ID,
					"created_at": bson.M{"$gte": cutoff},
				})
				if count > 0 {
					continue
				}

				msg := fmt.Sprintf("Task \"%s\" is due in %d day(s)", card.Title, days)
				n := &models.Notification{
					ID:        primitive.NewObjectID(),
					UserID:    user.ID,
					Type:      "due_soon",
					Message:   msg,
					ProjectID: card.ProjectID,
					CardID:    card.ID,
					Read:      false,
					CreatedAt: now,
				}
				database.GetCollection("notifications").InsertOne(ctx, n)
			}
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	database.Connect()
	startDueDateReminder()

	app := fiber.New(fiber.Config{
		AppName: "FPMB API",
	})

	app.Use(logger.New(logger.Config{
		Next: func(c *fiber.Ctx) bool {
			return len(c.Path()) >= 5 && c.Path()[:5] == "/_app" ||
				c.Path() == "/favicon.ico" ||
				len(c.Path()) >= 7 && c.Path()[:7] == "/fonts/"
		},
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "message": "FPMB API is running"})
	})

	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Post("/refresh", handlers.RefreshToken)
	auth.Post("/logout", middleware.Protected(), handlers.Logout)

	// Public avatar/media routes (no auth needed for <img> tags)
	api.Get("/avatar/:userId", handlers.ServePublicAvatar)
	api.Get("/team-media/:teamId/:imageType", handlers.ServePublicTeamImage)

	users := api.Group("/users", middleware.Protected())
	users.Get("/me", handlers.GetMe)
	users.Put("/me", handlers.UpdateMe)
	users.Put("/me/password", handlers.ChangePassword)
	users.Get("/search", handlers.SearchUsers)
	users.Post("/me/avatar", handlers.UploadUserAvatar)
	users.Get("/me/avatar", handlers.ServeUserAvatar)
	users.Get("/me/files", handlers.ListUserFiles)
	users.Post("/me/files/folder", handlers.CreateUserFolder)
	users.Post("/me/files/upload", handlers.UploadUserFile)
	users.Get("/me/api-keys", handlers.ListAPIKeys)
	users.Post("/me/api-keys", handlers.CreateAPIKey)
	users.Delete("/me/api-keys/:keyId", handlers.RevokeAPIKey)

	teams := api.Group("/teams", middleware.Protected())
	teams.Get("/", handlers.ListTeams)
	teams.Post("/", handlers.CreateTeam)
	teams.Get("/:teamId", handlers.GetTeam)
	teams.Put("/:teamId", handlers.UpdateTeam)
	teams.Delete("/:teamId", handlers.DeleteTeam)
	teams.Get("/:teamId/members", handlers.ListTeamMembers)
	teams.Post("/:teamId/members/invite", handlers.InviteTeamMember)
	teams.Put("/:teamId/members/:userId", handlers.UpdateTeamMemberRole)
	teams.Delete("/:teamId/members/:userId", handlers.RemoveTeamMember)
	teams.Get("/:teamId/projects", handlers.ListTeamProjects)
	teams.Post("/:teamId/projects", handlers.CreateProject)
	teams.Get("/:teamId/events", handlers.ListTeamEvents)
	teams.Post("/:teamId/events", handlers.CreateTeamEvent)
	teams.Get("/:teamId/docs", handlers.ListDocs)
	teams.Post("/:teamId/docs", handlers.CreateDoc)
	teams.Get("/:teamId/files", handlers.ListTeamFiles)
	teams.Post("/:teamId/files/folder", handlers.CreateTeamFolder)
	teams.Post("/:teamId/files/upload", handlers.UploadTeamFile)
	teams.Post("/:teamId/avatar", handlers.UploadTeamAvatar)
	teams.Get("/:teamId/avatar", handlers.ServeTeamAvatar)
	teams.Post("/:teamId/banner", handlers.UploadTeamBanner)
	teams.Get("/:teamId/banner", handlers.ServeTeamBanner)
	teams.Get("/:teamId/chat", handlers.ListChatMessages)

	projects := api.Group("/projects", middleware.Protected())
	projects.Get("/", handlers.ListProjects)
	projects.Post("/", handlers.CreatePersonalProject)
	projects.Get("/:projectId", handlers.GetProject)
	projects.Put("/:projectId", handlers.UpdateProject)
	projects.Put("/:projectId/archive", handlers.ArchiveProject)
	projects.Delete("/:projectId", handlers.DeleteProject)
	projects.Get("/:projectId/members", handlers.ListProjectMembers)
	projects.Post("/:projectId/members", handlers.AddProjectMember)
	projects.Put("/:projectId/members/:userId", handlers.UpdateProjectMemberRole)
	projects.Delete("/:projectId/members/:userId", handlers.RemoveProjectMember)
	projects.Get("/:projectId/board", handlers.GetBoard)
	projects.Post("/:projectId/columns", handlers.CreateColumn)
	projects.Put("/:projectId/columns/:columnId", handlers.UpdateColumn)
	projects.Put("/:projectId/columns/:columnId/position", handlers.ReorderColumn)
	projects.Delete("/:projectId/columns/:columnId", handlers.DeleteColumn)
	projects.Post("/:projectId/columns/:columnId/cards", handlers.CreateCard)
	projects.Get("/:projectId/events", handlers.ListProjectEvents)
	projects.Post("/:projectId/events", handlers.CreateProjectEvent)
	projects.Get("/:projectId/files", handlers.ListFiles)
	projects.Post("/:projectId/files/folder", handlers.CreateFolder)
	projects.Post("/:projectId/files/upload", handlers.UploadFile)
	projects.Get("/:projectId/webhooks", handlers.ListWebhooks)
	projects.Post("/:projectId/webhooks", handlers.CreateWebhook)
	projects.Get("/:projectId/whiteboard", handlers.GetWhiteboard)
	projects.Put("/:projectId/whiteboard", handlers.SaveWhiteboard)

	cards := api.Group("/cards", middleware.Protected())
	cards.Put("/:cardId", handlers.UpdateCard)
	cards.Put("/:cardId/move", handlers.MoveCard)
	cards.Delete("/:cardId", handlers.DeleteCard)

	events := api.Group("/events", middleware.Protected())
	events.Put("/:eventId", handlers.UpdateEvent)
	events.Delete("/:eventId", handlers.DeleteEvent)

	notifications := api.Group("/notifications", middleware.Protected())
	notifications.Get("/", handlers.ListNotifications)
	notifications.Put("/read-all", handlers.MarkAllNotificationsRead)
	notifications.Put("/:notifId/read", handlers.MarkNotificationRead)
	notifications.Delete("/:notifId", handlers.DeleteNotification)

	docs := api.Group("/docs", middleware.Protected())
	docs.Get("/:docId", handlers.GetDoc)
	docs.Put("/:docId", handlers.UpdateDoc)
	docs.Delete("/:docId", handlers.DeleteDoc)

	files := api.Group("/files", middleware.Protected())
	files.Get("/:fileId/download", handlers.DownloadFile)
	files.Delete("/:fileId", handlers.DeleteFile)

	webhooks := api.Group("/webhooks", middleware.Protected())
	webhooks.Put("/:webhookId", handlers.UpdateWebhook)
	webhooks.Put("/:webhookId/toggle", handlers.ToggleWebhook)
	webhooks.Delete("/:webhookId", handlers.DeleteWebhook)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/whiteboard/:id", websocket.New(handlers.WhiteboardWS))
	app.Get("/ws/team/:id/chat", websocket.New(handlers.TeamChatWS))

	app.Static("/", "../build")
	app.Get("/*", func(c *fiber.Ctx) error {
		if len(c.Path()) > 4 && c.Path()[:4] == "/api" {
			return c.Status(404).JSON(fiber.Map{"error": "Not Found"})
		}
		return c.SendFile("../build/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
