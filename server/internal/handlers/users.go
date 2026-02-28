package handlers

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fpmb/server/internal/database"
	"github.com/fpmb/server/internal/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func GetMe(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	if err := database.GetCollection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

func UpdateMe(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	var body struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"updated_at": time.Now()}
	if body.Name != "" {
		update["name"] = body.Name
	}
	if body.Email != "" {
		update["email"] = body.Email
	}
	if body.AvatarURL != "" {
		update["avatar_url"] = body.AvatarURL
	}

	col := database.GetCollection("users")
	if _, err := col.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": update}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user"})
	}

	var user models.User
	if err := col.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch updated user"})
	}

	return c.JSON(user)
}

func SearchUsers(c *fiber.Ctx) error {
	q := c.Query("q")
	if len(q) < 1 {
		return c.JSON([]fiber.Map{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": bson.A{
			bson.M{"name": bson.M{"$regex": q, "$options": "i"}},
			bson.M{"email": bson.M{"$regex": q, "$options": "i"}},
		},
	}

	cursor, err := database.GetCollection("users").Find(ctx, filter, options.Find().SetLimit(10))
	if err != nil {
		return c.JSON([]fiber.Map{})
	}
	defer cursor.Close(ctx)

	var users []models.User
	cursor.All(ctx, &users)

	type UserResult struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	result := []UserResult{}
	for _, u := range users {
		result = append(result, UserResult{ID: u.ID.Hex(), Name: u.Name, Email: u.Email})
	}
	return c.JSON(result)
}

func UploadUserAvatar(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No file provided"})
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedImageExts[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid image type"})
	}

	dir := filepath.Join("../data/users", userID.Hex())
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("UploadUserAvatar MkdirAll error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create storage directory"})
	}

	existingGlob := filepath.Join(dir, "avatar.*")
	if matches, _ := filepath.Glob(existingGlob); len(matches) > 0 {
		for _, m := range matches {
			os.Remove(m)
		}
	}

	destPath := filepath.Join(dir, "avatar"+ext)
	if err := c.SaveFile(fh, destPath); err != nil {
		log.Printf("UploadUserAvatar SaveFile error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save image"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := database.GetCollection("users")
	if _, err := col.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{
		"avatar_url": "/api/avatar/" + userID.Hex(),
		"updated_at": time.Now(),
	}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user"})
	}

	var user models.User
	if err := col.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch updated user"})
	}

	return c.JSON(user)
}

func ServeUserAvatar(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	dir := filepath.Join("../data/users", userID.Hex())
	for ext := range allowedImageExts {
		p := filepath.Join(dir, "avatar"+ext)
		if _, err := os.Stat(p); err == nil {
			return c.SendFile(p)
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Avatar not found"})
}

func ServePublicAvatar(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if _, err := primitive.ObjectIDFromHex(userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	dir := filepath.Join("../data/users", userID)
	for ext := range allowedImageExts {
		p := filepath.Join(dir, "avatar"+ext)
		if _, err := os.Stat(p); err == nil {
			c.Set("Cache-Control", "public, max-age=3600")
			return c.SendFile(p)
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Avatar not found"})
}

func ChangePassword(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.CurrentPassword == "" || body.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "current_password and new_password are required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := database.GetCollection("users")
	var user models.User
	if err := col.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.CurrentPassword)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Current password is incorrect"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	if _, err := col.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{
		"password_hash": string(hash),
		"updated_at":    time.Now(),
	}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{"message": "Password updated successfully"})
}
