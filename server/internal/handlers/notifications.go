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

func createNotification(ctx context.Context, userID primitive.ObjectID, notifType, message string, projectID primitive.ObjectID, cardID primitive.ObjectID) {
	n := &models.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Type:      notifType,
		Message:   message,
		ProjectID: projectID,
		CardID:    cardID,
		Read:      false,
		CreatedAt: time.Now(),
	}
	database.GetCollection("notifications").InsertOne(ctx, n)
}

func ListNotifications(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	if c.Query("read") == "false" {
		filter["read"] = false
	}

	cursor, err := database.GetCollection("notifications").Find(ctx, filter,
		options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch notifications"})
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	cursor.All(ctx, &notifications)
	if notifications == nil {
		notifications = []models.Notification{}
	}
	return c.JSON(notifications)
}

func MarkNotificationRead(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	notifID, err := primitive.ObjectIDFromHex(c.Params("notifId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid notification ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database.GetCollection("notifications").UpdateOne(ctx,
		bson.M{"_id": notifID, "user_id": userID},
		bson.M{"$set": bson.M{"read": true}},
	)
	return c.JSON(fiber.Map{"message": "Marked as read"})
}

func MarkAllNotificationsRead(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database.GetCollection("notifications").UpdateMany(ctx,
		bson.M{"user_id": userID, "read": false},
		bson.M{"$set": bson.M{"read": true}},
	)
	return c.JSON(fiber.Map{"message": "All notifications marked as read"})
}

func DeleteNotification(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	notifID, err := primitive.ObjectIDFromHex(c.Params("notifId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid notification ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database.GetCollection("notifications").DeleteOne(ctx, bson.M{"_id": notifID, "user_id": userID})
	return c.JSON(fiber.Map{"message": "Notification deleted"})
}
