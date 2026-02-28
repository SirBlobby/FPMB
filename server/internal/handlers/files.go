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

func storageBase(ctx context.Context, projectID primitive.ObjectID) (string, error) {
	var project models.Project
	if err := database.GetCollection("projects").FindOne(ctx, bson.M{"_id": projectID}).Decode(&project); err != nil {
		return "", err
	}
	if project.TeamID == primitive.NilObjectID {
		return filepath.Join("../data/users", project.CreatedBy.Hex()), nil
	}
	return filepath.Join("../data/teams", project.TeamID.Hex()), nil
}

func ListFiles(c *fiber.Ctx) error {
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
		log.Printf("ListFiles getProjectRole error: %v (projectID=%s userID=%s)", err, projectID.Hex(), userID.Hex())
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	filter := bson.M{"project_id": projectID}
	if parentID := c.Query("parent_id"); parentID != "" {
		oid, err := primitive.ObjectIDFromHex(parentID)
		if err == nil {
			filter["parent_id"] = oid
		}
	} else {
		filter["parent_id"] = bson.M{"$exists": false}
	}

	cursor, err := database.GetCollection("files").Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "type", Value: -1}, {Key: "name", Value: 1}}))
	if err != nil {
		log.Printf("ListFiles Find error: %v (filter=%v)", err, filter)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch files"})
	}
	defer cursor.Close(ctx)

	var files []models.File
	cursor.All(ctx, &files)
	if files == nil {
		files = []models.File{}
	}
	return c.JSON(files)
}

func CreateFolder(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	var body struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	now := time.Now()
	file := &models.File{
		ID:        primitive.NewObjectID(),
		ProjectID: projectID,
		Name:      body.Name,
		Type:      "folder",
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if body.ParentID != "" {
		if oid, err := primitive.ObjectIDFromHex(body.ParentID); err == nil {
			file.ParentID = &oid
		}
	}

	database.GetCollection("files").InsertOne(ctx, file)
	return c.Status(fiber.StatusCreated).JSON(file)
}

func UploadFile(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No file provided"})
	}

	base, err := storageBase(ctx, projectID)
	if err != nil {
		log.Printf("UploadFile storageBase error: %v (projectID=%s)", err, projectID.Hex())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to resolve storage path"})
	}

	if err := os.MkdirAll(base, 0755); err != nil {
		log.Printf("UploadFile MkdirAll error: %v (base=%s)", err, base)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create storage directory"})
	}

	filename := fh.Filename
	ext := filepath.Ext(filename)
	stem := filename[:len(filename)-len(ext)]
	destPath := filepath.Join(base, filename)
	for n := 2; ; n++ {
		if _, statErr := os.Stat(destPath); statErr != nil {
			break
		}
		filename = fmt.Sprintf("%s (%d)%s", stem, n, ext)
		destPath = filepath.Join(base, filename)
	}

	if err := c.SaveFile(fh, destPath); err != nil {
		log.Printf("UploadFile SaveFile error: %v (destPath=%s)", err, destPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	relPath := destPath[len("../data/"):]

	now := time.Now()
	file := &models.File{
		ID:         primitive.NewObjectID(),
		ProjectID:  projectID,
		Name:       fh.Filename,
		Type:       "file",
		SizeBytes:  fh.Size,
		StorageURL: relPath,
		CreatedBy:  userID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	parentID := c.FormValue("parent_id")
	if parentID != "" {
		if oid, err := primitive.ObjectIDFromHex(parentID); err == nil {
			file.ParentID = &oid
		}
	}

	database.GetCollection("files").InsertOne(ctx, file)
	return c.Status(fiber.StatusCreated).JSON(file)
}

func ListTeamFiles(c *fiber.Ctx) error {
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

	filter := bson.M{"team_id": teamID}
	if parentID := c.Query("parent_id"); parentID != "" {
		oid, err := primitive.ObjectIDFromHex(parentID)
		if err == nil {
			filter["parent_id"] = oid
		}
	} else {
		filter["parent_id"] = bson.M{"$exists": false}
	}

	cursor, err := database.GetCollection("files").Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "type", Value: -1}, {Key: "name", Value: 1}}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch files"})
	}
	defer cursor.Close(ctx)

	var files []models.File
	cursor.All(ctx, &files)
	if files == nil {
		files = []models.File{}
	}
	return c.JSON(files)
}

func CreateTeamFolder(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	teamID, err := primitive.ObjectIDFromHex(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	var body struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getTeamRole(ctx, teamID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	now := time.Now()
	file := &models.File{
		ID:        primitive.NewObjectID(),
		TeamID:    teamID,
		Name:      body.Name,
		Type:      "folder",
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if body.ParentID != "" {
		if oid, err := primitive.ObjectIDFromHex(body.ParentID); err == nil {
			file.ParentID = &oid
		}
	}

	database.GetCollection("files").InsertOne(ctx, file)
	return c.Status(fiber.StatusCreated).JSON(file)
}

func UploadTeamFile(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	teamID, err := primitive.ObjectIDFromHex(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	roleFlags, err := getTeamRole(ctx, teamID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No file provided"})
	}

	base := filepath.Join("../data/teams", teamID.Hex(), "files")
	if err := os.MkdirAll(base, 0755); err != nil {
		log.Printf("UploadTeamFile MkdirAll error: %v (base=%s)", err, base)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create storage directory"})
	}

	filename := fh.Filename
	ext := filepath.Ext(filename)
	stem := filename[:len(filename)-len(ext)]
	destPath := filepath.Join(base, filename)
	for n := 2; ; n++ {
		if _, statErr := os.Stat(destPath); statErr != nil {
			break
		}
		filename = fmt.Sprintf("%s (%d)%s", stem, n, ext)
		destPath = filepath.Join(base, filename)
	}

	if err := c.SaveFile(fh, destPath); err != nil {
		log.Printf("UploadTeamFile SaveFile error: %v (destPath=%s)", err, destPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	relPath := destPath[len("../data/"):]

	now := time.Now()
	file := &models.File{
		ID:         primitive.NewObjectID(),
		TeamID:     teamID,
		Name:       fh.Filename,
		Type:       "file",
		SizeBytes:  fh.Size,
		StorageURL: relPath,
		CreatedBy:  userID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	parentID := c.FormValue("parent_id")
	if parentID != "" {
		if oid, err := primitive.ObjectIDFromHex(parentID); err == nil {
			file.ParentID = &oid
		}
	}

	database.GetCollection("files").InsertOne(ctx, file)
	return c.Status(fiber.StatusCreated).JSON(file)
}

func ListUserFiles(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	if parentID := c.Query("parent_id"); parentID != "" {
		oid, err := primitive.ObjectIDFromHex(parentID)
		if err == nil {
			filter["parent_id"] = oid
		}
	} else {
		filter["parent_id"] = bson.M{"$exists": false}
	}

	cursor, err := database.GetCollection("files").Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "type", Value: -1}, {Key: "name", Value: 1}}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch files"})
	}
	defer cursor.Close(ctx)

	var files []models.File
	cursor.All(ctx, &files)
	if files == nil {
		files = []models.File{}
	}
	return c.JSON(files)
}

func CreateUserFolder(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	var body struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	now := time.Now()
	file := &models.File{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Name:      body.Name,
		Type:      "folder",
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if body.ParentID != "" {
		if oid, err := primitive.ObjectIDFromHex(body.ParentID); err == nil {
			file.ParentID = &oid
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database.GetCollection("files").InsertOne(ctx, file)
	return c.Status(fiber.StatusCreated).JSON(file)
}

func UploadUserFile(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No file provided"})
	}

	base := filepath.Join("../data/users", userID.Hex(), "files")
	if err := os.MkdirAll(base, 0755); err != nil {
		log.Printf("UploadUserFile MkdirAll error: %v (base=%s)", err, base)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create storage directory"})
	}

	filename := fh.Filename
	ext := filepath.Ext(filename)
	stem := filename[:len(filename)-len(ext)]
	destPath := filepath.Join(base, filename)
	for n := 2; ; n++ {
		if _, statErr := os.Stat(destPath); statErr != nil {
			break
		}
		filename = fmt.Sprintf("%s (%d)%s", stem, n, ext)
		destPath = filepath.Join(base, filename)
	}

	if err := c.SaveFile(fh, destPath); err != nil {
		log.Printf("UploadUserFile SaveFile error: %v (destPath=%s)", err, destPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	relPath := destPath[len("../data/"):]

	now := time.Now()
	file := &models.File{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		Name:       fh.Filename,
		Type:       "file",
		SizeBytes:  fh.Size,
		StorageURL: relPath,
		CreatedBy:  userID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	parentID := c.FormValue("parent_id")
	if parentID != "" {
		if oid, err := primitive.ObjectIDFromHex(parentID); err == nil {
			file.ParentID = &oid
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database.GetCollection("files").InsertOne(ctx, file)
	return c.Status(fiber.StatusCreated).JSON(file)
}

func DownloadFile(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	fileID, err := primitive.ObjectIDFromHex(c.Params("fileId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var file models.File
	if err := database.GetCollection("files").FindOne(ctx, bson.M{"_id": fileID}).Decode(&file); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	if file.TeamID != primitive.NilObjectID {
		if _, err := getTeamRole(ctx, file.TeamID, userID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
		}
	} else if file.UserID != primitive.NilObjectID {
		if file.UserID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
		}
	} else {
		if _, err := getProjectRole(ctx, file.ProjectID, userID); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
		}
	}

	if file.Type == "folder" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot download a folder"})
	}

	diskPath := filepath.Join("../data", file.StorageURL)
	if _, err := os.Stat(diskPath); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found on disk"})
	}

	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Name))
	return c.SendFile(diskPath)
}

func DeleteFile(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	fileID, err := primitive.ObjectIDFromHex(c.Params("fileId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var file models.File
	if err := database.GetCollection("files").FindOne(ctx, bson.M{"_id": fileID}).Decode(&file); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	var roleFlags int
	if file.TeamID != primitive.NilObjectID {
		roleFlags, err = getTeamRole(ctx, file.TeamID, userID)
	} else if file.UserID != primitive.NilObjectID {
		if file.UserID == userID {
			roleFlags = RoleOwner
		} else {
			err = fmt.Errorf("access denied")
		}
	} else {
		roleFlags, err = getProjectRole(ctx, file.ProjectID, userID)
	}
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("files").DeleteOne(ctx, bson.M{"_id": fileID})

	if file.Type == "folder" {
		database.GetCollection("files").DeleteMany(ctx, bson.M{"parent_id": fileID})
	} else if file.StorageURL != "" {
		os.Remove(filepath.Join("../data", file.StorageURL))
	}

	return c.JSON(fiber.Map{"message": "Deleted"})
}
