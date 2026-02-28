package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/fpmb/server/internal/database"
	"github.com/fpmb/server/internal/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// generateAPIKey returns a 32-byte random hex token (64 chars) prefixed with "fpmb_".
func generateAPIKey() (raw string, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = "fpmb_" + hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hashed = hex.EncodeToString(sum[:])
	return
}

// ListAPIKeys returns all non-revoked API keys for the current user (without exposing hashes).
func ListAPIKeys(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.GetCollection("api_keys").Find(ctx, bson.M{
		"user_id":    userID,
		"revoked_at": bson.M{"$exists": false},
	})
	if err != nil {
		return c.JSON([]fiber.Map{})
	}
	defer cursor.Close(ctx)

	var keys []models.APIKey
	cursor.All(ctx, &keys)

	// Strip the hash before returning.
	type SafeKey struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Scopes    []string  `json:"scopes"`
		Prefix    string    `json:"prefix"`
		LastUsed  *time.Time `json:"last_used,omitempty"`
		CreatedAt time.Time `json:"created_at"`
	}
	result := []SafeKey{}
	for _, k := range keys {
		result = append(result, SafeKey{
			ID:        k.ID.Hex(),
			Name:      k.Name,
			Scopes:    k.Scopes,
			Prefix:    k.Prefix,
			LastUsed:  k.LastUsed,
			CreatedAt: k.CreatedAt,
		})
	}
	return c.JSON(result)
}

// CreateAPIKey generates a new API key and stores its hash.
func CreateAPIKey(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	var body struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	if len(body.Scopes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "at least one scope is required"})
	}

	// Validate scopes.
	valid := map[string]bool{
		"read:projects": true, "write:projects": true,
		"read:boards": true, "write:boards": true,
		"read:teams": true, "write:teams": true,
		"read:files": true, "write:files": true,
		"read:notifications": true,
	}
	for _, s := range body.Scopes {
		if !valid[s] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unknown scope: " + s})
		}
	}

	raw, hashed, err := generateAPIKey()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate key"})
	}

	now := time.Now()
	key := models.APIKey{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Name:      body.Name,
		Scopes:    body.Scopes,
		KeyHash:   hashed,
		Prefix:    raw[:10], // "fpmb_" + first 5 chars of random
		CreatedAt: now,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := database.GetCollection("api_keys").InsertOne(ctx, key); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to store key"})
	}

	// Return the raw key only once.
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         key.ID.Hex(),
		"name":       key.Name,
		"scopes":     key.Scopes,
		"prefix":     key.Prefix,
		"key":        raw,
		"created_at": key.CreatedAt,
	})
}

// RevokeAPIKey soft-deletes an API key belonging to the current user.
func RevokeAPIKey(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	keyID, err := primitive.ObjectIDFromHex(c.Params("keyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid key ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	res, err := database.GetCollection("api_keys").UpdateOne(ctx,
		bson.M{"_id": keyID, "user_id": userID},
		bson.M{"$set": bson.M{"revoked_at": now}},
	)
	if err != nil || res.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Key not found"})
	}

	return c.JSON(fiber.Map{"message": "Key revoked"})
}
