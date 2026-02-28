package handlers

import (
	"context"
	"time"

	"github.com/fpmb/server/internal/database"
	"github.com/fpmb/server/internal/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListTeamEvents(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	teamID, err := primitive.ObjectIDFromHex(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := getTeamRole(ctx, teamID, userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	filter := bson.M{"scope_id": teamID, "scope": "org"}
	if month := c.Query("month"); month != "" {
		filter["date"] = bson.M{"$regex": "^" + month}
	}

	cursor, err := database.GetCollection("events").Find(ctx, filter,
		options.Find().SetSort(bson.M{"date": 1}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch events"})
	}
	defer cursor.Close(ctx)

	var events []models.Event
	cursor.All(ctx, &events)
	if events == nil {
		events = []models.Event{}
	}
	return c.JSON(events)
}

func CreateTeamEvent(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	teamID, err := primitive.ObjectIDFromHex(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	var body struct {
		Title       string `json:"title"`
		Date        string `json:"date"`
		Time        string `json:"time"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil || body.Title == "" || body.Date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title and date are required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getTeamRole(ctx, teamID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	now := time.Now()
	event := &models.Event{
		ID:          primitive.NewObjectID(),
		Title:       body.Title,
		Date:        body.Date,
		Time:        body.Time,
		Color:       body.Color,
		Description: body.Description,
		Scope:       "org",
		ScopeID:     teamID,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	database.GetCollection("events").InsertOne(ctx, event)
	return c.Status(fiber.StatusCreated).JSON(event)
}

func ListProjectEvents(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := getProjectRole(ctx, projectID, userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	filter := bson.M{"scope_id": projectID, "scope": "project"}
	if month := c.Query("month"); month != "" {
		filter["date"] = bson.M{"$regex": "^" + month}
	}

	cursor, err := database.GetCollection("events").Find(ctx, filter,
		options.Find().SetSort(bson.M{"date": 1}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch events"})
	}
	defer cursor.Close(ctx)

	var events []models.Event
	cursor.All(ctx, &events)
	if events == nil {
		events = []models.Event{}
	}
	return c.JSON(events)
}

func CreateProjectEvent(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	var body struct {
		Title       string `json:"title"`
		Date        string `json:"date"`
		Time        string `json:"time"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil || body.Title == "" || body.Date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title and date are required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	now := time.Now()
	event := &models.Event{
		ID:          primitive.NewObjectID(),
		Title:       body.Title,
		Date:        body.Date,
		Time:        body.Time,
		Color:       body.Color,
		Description: body.Description,
		Scope:       "project",
		ScopeID:     projectID,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	database.GetCollection("events").InsertOne(ctx, event)
	return c.Status(fiber.StatusCreated).JSON(event)
}

func UpdateEvent(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	eventID, err := primitive.ObjectIDFromHex(c.Params("eventId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var event models.Event
	if err := database.GetCollection("events").FindOne(ctx, bson.M{"_id": eventID}).Decode(&event); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	var roleFlags int
	var roleErr error
	if event.Scope == "org" {
		roleFlags, roleErr = getTeamRole(ctx, event.ScopeID, userID)
	} else {
		roleFlags, roleErr = getProjectRole(ctx, event.ScopeID, userID)
	}
	if roleErr != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var body struct {
		Title       string `json:"title"`
		Date        string `json:"date"`
		Time        string `json:"time"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}
	c.BodyParser(&body)

	update := bson.M{"updated_at": time.Now()}
	if body.Title != "" {
		update["title"] = body.Title
	}
	if body.Date != "" {
		update["date"] = body.Date
	}
	if body.Time != "" {
		update["time"] = body.Time
	}
	if body.Color != "" {
		update["color"] = body.Color
	}
	if body.Description != "" {
		update["description"] = body.Description
	}

	col := database.GetCollection("events")
	col.UpdateOne(ctx, bson.M{"_id": eventID}, bson.M{"$set": update})

	var updated models.Event
	col.FindOne(ctx, bson.M{"_id": eventID}).Decode(&updated)
	return c.JSON(updated)
}

func DeleteEvent(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	eventID, err := primitive.ObjectIDFromHex(c.Params("eventId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var event models.Event
	if err := database.GetCollection("events").FindOne(ctx, bson.M{"_id": eventID}).Decode(&event); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	var roleFlags int
	var roleErr error
	if event.Scope == "org" {
		roleFlags, roleErr = getTeamRole(ctx, event.ScopeID, userID)
	} else {
		roleFlags, roleErr = getProjectRole(ctx, event.ScopeID, userID)
	}
	if roleErr != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("events").DeleteOne(ctx, bson.M{"_id": eventID})
	return c.JSON(fiber.Map{"message": "Event deleted"})
}
