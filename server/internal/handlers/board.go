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

func GetBoard(c *fiber.Ctx) error {
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

	colCursor, err := database.GetCollection("board_columns").Find(ctx,
		bson.M{"project_id": projectID},
		options.Find().SetSort(bson.M{"position": 1}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch columns"})
	}
	defer colCursor.Close(ctx)

	var columns []models.BoardColumn
	colCursor.All(ctx, &columns)

	type ColumnWithCards struct {
		models.BoardColumn
		Cards []models.Card `json:"cards"`
	}

	result := []ColumnWithCards{}
	for _, col := range columns {
		cardCursor, err := database.GetCollection("cards").Find(ctx,
			bson.M{"column_id": col.ID},
			options.Find().SetSort(bson.M{"position": 1}))
		if err != nil {
			result = append(result, ColumnWithCards{BoardColumn: col, Cards: []models.Card{}})
			continue
		}
		var cards []models.Card
		cardCursor.All(ctx, &cards)
		cardCursor.Close(ctx)
		if cards == nil {
			cards = []models.Card{}
		}
		result = append(result, ColumnWithCards{BoardColumn: col, Cards: cards})
	}

	return c.JSON(fiber.Map{"project_id": projectID, "columns": result})
}

func CreateColumn(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := c.BodyParser(&body); err != nil || body.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	count, _ := database.GetCollection("board_columns").CountDocuments(ctx, bson.M{"project_id": projectID})
	now := time.Now()
	col := &models.BoardColumn{
		ID:        primitive.NewObjectID(),
		ProjectID: projectID,
		Title:     body.Title,
		Position:  int(count),
		CreatedAt: now,
		UpdatedAt: now,
	}

	database.GetCollection("board_columns").InsertOne(ctx, col)
	return c.Status(fiber.StatusCreated).JSON(col)
}

func UpdateColumn(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	columnID, err := primitive.ObjectIDFromHex(c.Params("columnId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid column ID"})
	}

	var body struct {
		Title string `json:"title"`
	}
	c.BodyParser(&body)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	update := bson.M{"updated_at": time.Now()}
	if body.Title != "" {
		update["title"] = body.Title
	}

	col := database.GetCollection("board_columns")
	col.UpdateOne(ctx, bson.M{"_id": columnID, "project_id": projectID}, bson.M{"$set": update})

	var column models.BoardColumn
	col.FindOne(ctx, bson.M{"_id": columnID}).Decode(&column)
	return c.JSON(column)
}

func ReorderColumn(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	columnID, err := primitive.ObjectIDFromHex(c.Params("columnId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid column ID"})
	}

	var body struct {
		Position int `json:"position"`
	}
	c.BodyParser(&body)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("board_columns").UpdateOne(ctx,
		bson.M{"_id": columnID, "project_id": projectID},
		bson.M{"$set": bson.M{"position": body.Position, "updated_at": time.Now()}},
	)
	return c.JSON(fiber.Map{"id": columnID, "position": body.Position})
}

func DeleteColumn(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	columnID, err := primitive.ObjectIDFromHex(c.Params("columnId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid column ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("board_columns").DeleteOne(ctx, bson.M{"_id": columnID, "project_id": projectID})
	database.GetCollection("cards").DeleteMany(ctx, bson.M{"column_id": columnID})
	return c.JSON(fiber.Map{"message": "Column deleted"})
}

func CreateCard(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	columnID, err := primitive.ObjectIDFromHex(c.Params("columnId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid column ID"})
	}

	var body struct {
		Title       string           `json:"title"`
		Description string           `json:"description"`
		Priority    string           `json:"priority"`
		Color       string           `json:"color"`
		DueDate     string           `json:"due_date"`
		Assignees   []string         `json:"assignees"`
		Subtasks    []models.Subtask `json:"subtasks"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	count, _ := database.GetCollection("cards").CountDocuments(ctx, bson.M{"column_id": columnID})
	now := time.Now()

	if body.Assignees == nil {
		body.Assignees = []string{}
	}
	if body.Subtasks == nil {
		body.Subtasks = []models.Subtask{}
	}
	if body.Priority == "" {
		body.Priority = "Medium"
	}
	if body.Color == "" {
		body.Color = "neutral"
	}

	var dueDate *time.Time
	if body.DueDate != "" {
		if parsed, parseErr := time.Parse("2006-01-02", body.DueDate); parseErr == nil {
			dueDate = &parsed
		}
	}

	card := &models.Card{
		ID:          primitive.NewObjectID(),
		ColumnID:    columnID,
		ProjectID:   projectID,
		Title:       body.Title,
		Description: body.Description,
		Priority:    body.Priority,
		Color:       body.Color,
		DueDate:     dueDate,
		Assignees:   body.Assignees,
		Subtasks:    body.Subtasks,
		Position:    int(count),
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	database.GetCollection("cards").InsertOne(ctx, card)

	for _, email := range card.Assignees {
		var assignee models.User
		if err := database.GetCollection("users").FindOne(ctx, bson.M{"email": email}).Decode(&assignee); err != nil {
			continue
		}
		if assignee.ID == userID {
			continue
		}
		createNotification(ctx, assignee.ID, "assign",
			"You have been assigned to the task \""+card.Title+"\"",
			card.ProjectID, card.ID)
	}

	return c.Status(fiber.StatusCreated).JSON(card)
}

func UpdateCard(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	cardID, err := primitive.ObjectIDFromHex(c.Params("cardId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid card ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existing models.Card
	if err := database.GetCollection("cards").FindOne(ctx, bson.M{"_id": cardID}).Decode(&existing); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Card not found"})
	}

	roleFlags, err := getProjectRole(ctx, existing.ProjectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var body struct {
		Title       *string          `json:"title"`
		Description *string          `json:"description"`
		Priority    *string          `json:"priority"`
		Color       *string          `json:"color"`
		DueDate     *string          `json:"due_date"`
		Assignees   []string         `json:"assignees"`
		Subtasks    []models.Subtask `json:"subtasks"`
	}
	c.BodyParser(&body)

	update := bson.M{"updated_at": time.Now()}
	if body.Title != nil {
		update["title"] = *body.Title
	}
	if body.Description != nil {
		update["description"] = *body.Description
	}
	if body.Priority != nil {
		update["priority"] = *body.Priority
	}
	if body.Color != nil {
		update["color"] = *body.Color
	}
	if body.DueDate != nil {
		if *body.DueDate == "" {
			update["due_date"] = nil
		} else if parsed, parseErr := time.Parse("2006-01-02", *body.DueDate); parseErr == nil {
			update["due_date"] = parsed
		}
	}
	if body.Assignees != nil {
		update["assignees"] = body.Assignees
	}
	if body.Subtasks != nil {
		update["subtasks"] = body.Subtasks
	}

	col := database.GetCollection("cards")
	col.UpdateOne(ctx, bson.M{"_id": cardID}, bson.M{"$set": update})

	var card models.Card
	col.FindOne(ctx, bson.M{"_id": cardID}).Decode(&card)

	if body.Assignees != nil {
		existingSet := make(map[string]bool)
		for _, e := range existing.Assignees {
			existingSet[e] = true
		}
		for _, email := range body.Assignees {
			if existingSet[email] {
				continue
			}
			var assignee models.User
			if err := database.GetCollection("users").FindOne(ctx, bson.M{"email": email}).Decode(&assignee); err != nil {
				continue
			}
			if assignee.ID == userID {
				continue
			}
			createNotification(ctx, assignee.ID, "assign",
				"You have been assigned to the task \""+card.Title+"\"",
				card.ProjectID, card.ID)
		}
	}

	return c.JSON(card)
}

func MoveCard(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	cardID, err := primitive.ObjectIDFromHex(c.Params("cardId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid card ID"})
	}

	var body struct {
		ColumnID string `json:"column_id"`
		Position int    `json:"position"`
	}
	if err := c.BodyParser(&body); err != nil || body.ColumnID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "column_id is required"})
	}

	newColumnID, err := primitive.ObjectIDFromHex(body.ColumnID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid column_id"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var card models.Card
	if err := database.GetCollection("cards").FindOne(ctx, bson.M{"_id": cardID}).Decode(&card); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Card not found"})
	}

	roleFlags, err := getProjectRole(ctx, card.ProjectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	col := database.GetCollection("cards")
	col.UpdateOne(ctx, bson.M{"_id": cardID}, bson.M{"$set": bson.M{
		"column_id":  newColumnID,
		"position":   body.Position,
		"updated_at": time.Now(),
	}})

	var updated models.Card
	col.FindOne(ctx, bson.M{"_id": cardID}).Decode(&updated)
	return c.JSON(updated)
}

func DeleteCard(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	cardID, err := primitive.ObjectIDFromHex(c.Params("cardId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid card ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var card models.Card
	if err := database.GetCollection("cards").FindOne(ctx, bson.M{"_id": cardID}).Decode(&card); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Card not found"})
	}

	roleFlags, err := getProjectRole(ctx, card.ProjectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("cards").DeleteOne(ctx, bson.M{"_id": cardID})
	return c.JSON(fiber.Map{"message": "Card deleted"})
}
