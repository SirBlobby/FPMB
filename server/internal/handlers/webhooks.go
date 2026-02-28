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

func ListWebhooks(c *fiber.Ctx) error {
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

	cursor, err := database.GetCollection("webhooks").Find(ctx,
		bson.M{"project_id": projectID},
		options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch webhooks"})
	}
	defer cursor.Close(ctx)

	var webhooks []models.Webhook
	cursor.All(ctx, &webhooks)
	if webhooks == nil {
		webhooks = []models.Webhook{}
	}
	return c.JSON(webhooks)
}

func CreateWebhook(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	var body struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		URL    string `json:"url"`
		Secret string `json:"secret"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" || body.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and url are required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	wType := body.Type
	if wType == "" {
		wType = "custom"
	}

	now := time.Now()
	webhook := &models.Webhook{
		ID:        primitive.NewObjectID(),
		ProjectID: projectID,
		Name:      body.Name,
		Type:      wType,
		URL:       body.URL,
		Status:    "active",
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if body.Secret != "" {
		webhook.SecretHash = body.Secret
	}

	database.GetCollection("webhooks").InsertOne(ctx, webhook)
	return c.Status(fiber.StatusCreated).JSON(webhook)
}

func UpdateWebhook(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	webhookID, err := primitive.ObjectIDFromHex(c.Params("webhookId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wh models.Webhook
	if err := database.GetCollection("webhooks").FindOne(ctx, bson.M{"_id": webhookID}).Decode(&wh); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Webhook not found"})
	}

	roleFlags, err := getProjectRole(ctx, wh.ProjectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var body struct {
		Name string `json:"name"`
		URL  string `json:"url"`
		Type string `json:"type"`
	}
	c.BodyParser(&body)

	update := bson.M{"updated_at": time.Now()}
	if body.Name != "" {
		update["name"] = body.Name
	}
	if body.URL != "" {
		update["url"] = body.URL
	}
	if body.Type != "" {
		update["type"] = body.Type
	}

	col := database.GetCollection("webhooks")
	col.UpdateOne(ctx, bson.M{"_id": webhookID}, bson.M{"$set": update})

	var updated models.Webhook
	col.FindOne(ctx, bson.M{"_id": webhookID}).Decode(&updated)
	return c.JSON(updated)
}

func ToggleWebhook(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	webhookID, err := primitive.ObjectIDFromHex(c.Params("webhookId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wh models.Webhook
	if err := database.GetCollection("webhooks").FindOne(ctx, bson.M{"_id": webhookID}).Decode(&wh); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Webhook not found"})
	}

	roleFlags, err := getProjectRole(ctx, wh.ProjectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	newStatus := "active"
	if wh.Status == "active" {
		newStatus = "inactive"
	}

	col := database.GetCollection("webhooks")
	col.UpdateOne(ctx, bson.M{"_id": webhookID}, bson.M{"$set": bson.M{
		"status":     newStatus,
		"updated_at": time.Now(),
	}})

	var updated models.Webhook
	col.FindOne(ctx, bson.M{"_id": webhookID}).Decode(&updated)
	return c.JSON(updated)
}

func DeleteWebhook(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	webhookID, err := primitive.ObjectIDFromHex(c.Params("webhookId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wh models.Webhook
	if err := database.GetCollection("webhooks").FindOne(ctx, bson.M{"_id": webhookID}).Decode(&wh); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Webhook not found"})
	}

	roleFlags, err := getProjectRole(ctx, wh.ProjectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("webhooks").DeleteOne(ctx, bson.M{"_id": webhookID})
	return c.JSON(fiber.Map{"message": "Webhook deleted"})
}
