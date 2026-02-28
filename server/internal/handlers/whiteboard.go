package handlers

import (
	"context"
	"time"

	"github.com/fpmb/server/internal/database"
	"github.com/fpmb/server/internal/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetWhiteboard(c *fiber.Ctx) error {
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

	var wb models.Whiteboard
	err = database.GetCollection("whiteboards").FindOne(ctx, bson.M{"project_id": projectID}).Decode(&wb)
	if err == mongo.ErrNoDocuments {
		return c.JSON(fiber.Map{"id": nil, "project_id": projectID, "data": "", "updated_at": nil})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch whiteboard"})
	}

	return c.JSON(wb)
}

func SaveWhiteboard(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	var body struct {
		Data string `json:"data"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	now := time.Now()
	col := database.GetCollection("whiteboards")

	var existing models.Whiteboard
	err = col.FindOne(ctx, bson.M{"project_id": projectID}).Decode(&existing)

	if err == mongo.ErrNoDocuments {
		wb := &models.Whiteboard{
			ID:        primitive.NewObjectID(),
			ProjectID: projectID,
			Data:      body.Data,
			CreatedBy: userID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		col.InsertOne(ctx, wb)
		return c.JSON(fiber.Map{"id": wb.ID, "project_id": projectID, "updated_at": now})
	}

	col.UpdateOne(ctx, bson.M{"project_id": projectID}, bson.M{"$set": bson.M{
		"data":       body.Data,
		"updated_at": now,
	}})

	return c.JSON(fiber.Map{"id": existing.ID, "project_id": projectID, "updated_at": now})
}
