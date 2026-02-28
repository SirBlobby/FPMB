package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fpmb/server/internal/database"
	"github.com/fpmb/server/internal/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListDocs(c *fiber.Ctx) error {
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

	cursor, err := database.GetCollection("docs").Find(ctx,
		bson.M{"team_id": teamID},
		options.Find().SetSort(bson.M{"updated_at": -1}).SetProjection(bson.M{"content": 0}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch docs"})
	}
	defer cursor.Close(ctx)

	var docs []models.Doc
	cursor.All(ctx, &docs)
	if docs == nil {
		docs = []models.Doc{}
	}
	return c.JSON(docs)
}

func CreateDoc(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	teamID, err := primitive.ObjectIDFromHex(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil || body.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getTeamRole(ctx, teamID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	now := time.Now()
	doc := &models.Doc{
		ID:        primitive.NewObjectID(),
		TeamID:    teamID,
		Title:     body.Title,
		Content:   body.Content,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	database.GetCollection("docs").InsertOne(ctx, doc)

	docDir := filepath.Join("../data/teams", teamID.Hex(), "docs")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		log.Printf("CreateDoc: mkdir %s: %v", docDir, err)
	} else {
		content := fmt.Sprintf("# %s\n\n%s", doc.Title, doc.Content)
		if err := os.WriteFile(filepath.Join(docDir, doc.ID.Hex()+".md"), []byte(content), 0644); err != nil {
			log.Printf("CreateDoc: write file: %v", err)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(doc)
}

func GetDoc(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	docID, err := primitive.ObjectIDFromHex(c.Params("docId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid doc ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var doc models.Doc
	if err := database.GetCollection("docs").FindOne(ctx, bson.M{"_id": docID}).Decode(&doc); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Doc not found"})
	}

	if _, err := getTeamRole(ctx, doc.TeamID, userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	return c.JSON(doc)
}

func UpdateDoc(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	docID, err := primitive.ObjectIDFromHex(c.Params("docId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid doc ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existing models.Doc
	if err := database.GetCollection("docs").FindOne(ctx, bson.M{"_id": docID}).Decode(&existing); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Doc not found"})
	}

	roleFlags, err := getTeamRole(ctx, existing.TeamID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	c.BodyParser(&body)

	update := bson.M{"updated_at": time.Now()}
	if body.Title != "" {
		update["title"] = body.Title
	}
	if body.Content != "" {
		update["content"] = body.Content
	}

	col := database.GetCollection("docs")
	col.UpdateOne(ctx, bson.M{"_id": docID}, bson.M{"$set": update})

	var doc models.Doc
	col.FindOne(ctx, bson.M{"_id": docID}).Decode(&doc)

	docDir := filepath.Join("../data/teams", existing.TeamID.Hex(), "docs")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		log.Printf("UpdateDoc: mkdir %s: %v", docDir, err)
	} else {
		content := fmt.Sprintf("# %s\n\n%s", doc.Title, doc.Content)
		if err := os.WriteFile(filepath.Join(docDir, doc.ID.Hex()+".md"), []byte(content), 0644); err != nil {
			log.Printf("UpdateDoc: write file: %v", err)
		}
	}

	return c.JSON(doc)
}

func DeleteDoc(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	docID, err := primitive.ObjectIDFromHex(c.Params("docId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid doc ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var doc models.Doc
	if err := database.GetCollection("docs").FindOne(ctx, bson.M{"_id": docID}).Decode(&doc); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Doc not found"})
	}

	roleFlags, err := getTeamRole(ctx, doc.TeamID, userID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("docs").DeleteOne(ctx, bson.M{"_id": docID})

	mdPath := filepath.Join("../data/teams", doc.TeamID.Hex(), "docs", docID.Hex()+".md")
	if err := os.Remove(mdPath); err != nil && !os.IsNotExist(err) {
		log.Printf("DeleteDoc: remove file %s: %v", mdPath, err)
	}

	return c.JSON(fiber.Map{"message": "Doc deleted"})
}
