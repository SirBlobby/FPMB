package handlers

import (
	"context"
	"fmt"
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
)

const (
	RoleViewer = 1
	RoleEditor = 2
	RoleAdmin  = 4
	RoleOwner  = 8
)

func hasPermission(userRole, requiredRole int) bool {
	return userRole >= requiredRole
}

func roleName(flags int) string {
	switch {
	case flags&RoleOwner != 0:
		return "Owner"
	case flags&RoleAdmin != 0:
		return "Admin"
	case flags&RoleEditor != 0:
		return "Editor"
	default:
		return "Viewer"
	}
}

func getTeamRole(ctx context.Context, teamID, userID primitive.ObjectID) (int, error) {
	var member models.TeamMember
	err := database.GetCollection("team_members").FindOne(ctx, bson.M{
		"team_id": teamID,
		"user_id": userID,
	}).Decode(&member)
	if err != nil {
		return 0, err
	}
	return member.RoleFlags, nil
}

func ListTeams(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.GetCollection("team_members").Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch teams"})
	}
	defer cursor.Close(ctx)

	var memberships []models.TeamMember
	if err := cursor.All(ctx, &memberships); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode memberships"})
	}

	type TeamResponse struct {
		ID          primitive.ObjectID `json:"id"`
		Name        string             `json:"name"`
		WorkspaceID string             `json:"workspace_id"`
		MemberCount int64              `json:"member_count"`
		RoleFlags   int                `json:"role_flags"`
		RoleName    string             `json:"role_name"`
		CreatedAt   time.Time          `json:"created_at"`
	}

	result := []TeamResponse{}
	for _, m := range memberships {
		var team models.Team
		if err := database.GetCollection("teams").FindOne(ctx, bson.M{"_id": m.TeamID}).Decode(&team); err != nil {
			continue
		}
		count, _ := database.GetCollection("team_members").CountDocuments(ctx, bson.M{"team_id": m.TeamID})
		result = append(result, TeamResponse{
			ID:          team.ID,
			Name:        team.Name,
			WorkspaceID: team.WorkspaceID,
			MemberCount: count,
			RoleFlags:   m.RoleFlags,
			RoleName:    roleName(m.RoleFlags),
			CreatedAt:   team.CreatedAt,
		})
	}

	return c.JSON(result)
}

func CreateTeam(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	var body struct {
		Name        string `json:"name"`
		WorkspaceID string `json:"workspace_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Team name is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	team := &models.Team{
		ID:          primitive.NewObjectID(),
		Name:        body.Name,
		WorkspaceID: body.WorkspaceID,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if _, err := database.GetCollection("teams").InsertOne(ctx, team); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create team"})
	}

	member := &models.TeamMember{
		ID:        primitive.NewObjectID(),
		TeamID:    team.ID,
		UserID:    userID,
		RoleFlags: RoleOwner,
		InvitedBy: userID,
		JoinedAt:  now,
	}
	if _, err := database.GetCollection("team_members").InsertOne(ctx, member); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add team owner"})
	}

	return c.Status(fiber.StatusCreated).JSON(team)
}

func GetTeam(c *fiber.Ctx) error {
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

	roleFlags, err := getTeamRole(ctx, teamID, userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	var team models.Team
	if err := database.GetCollection("teams").FindOne(ctx, bson.M{"_id": teamID}).Decode(&team); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Team not found"})
	}

	count, _ := database.GetCollection("team_members").CountDocuments(ctx, bson.M{"team_id": teamID})

	return c.JSON(fiber.Map{
		"id":           team.ID,
		"name":         team.Name,
		"workspace_id": team.WorkspaceID,
		"avatar_url":   team.AvatarURL,
		"banner_url":   team.BannerURL,
		"member_count": count,
		"role_flags":   roleFlags,
		"role_name":    roleName(roleFlags),
		"created_at":   team.CreatedAt,
		"updated_at":   team.UpdatedAt,
	})
}

func UpdateTeam(c *fiber.Ctx) error {
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

	roleFlags, err := getTeamRole(ctx, teamID, userID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var body struct {
		Name        string `json:"name"`
		WorkspaceID string `json:"workspace_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	update := bson.M{"updated_at": time.Now()}
	if body.Name != "" {
		update["name"] = body.Name
	}
	if body.WorkspaceID != "" {
		update["workspace_id"] = body.WorkspaceID
	}

	col := database.GetCollection("teams")
	if _, err := col.UpdateOne(ctx, bson.M{"_id": teamID}, bson.M{"$set": update}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update team"})
	}

	var team models.Team
	col.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	return c.JSON(team)
}

func DeleteTeam(c *fiber.Ctx) error {
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

	roleFlags, err := getTeamRole(ctx, teamID, userID)
	if err != nil || !hasPermission(roleFlags, RoleOwner) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only owners can delete teams"})
	}

	database.GetCollection("teams").DeleteOne(ctx, bson.M{"_id": teamID})
	database.GetCollection("team_members").DeleteMany(ctx, bson.M{"team_id": teamID})

	return c.JSON(fiber.Map{"message": "Team deleted"})
}

func ListTeamMembers(c *fiber.Ctx) error {
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

	cursor, err := database.GetCollection("team_members").Find(ctx, bson.M{"team_id": teamID},
		options.Find().SetSort(bson.M{"joined_at": 1}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch members"})
	}
	defer cursor.Close(ctx)

	var memberships []models.TeamMember
	cursor.All(ctx, &memberships)

	type MemberResponse struct {
		ID        primitive.ObjectID `json:"id"`
		UserID    primitive.ObjectID `json:"user_id"`
		Name      string             `json:"name"`
		Email     string             `json:"email"`
		RoleFlags int                `json:"role_flags"`
		RoleName  string             `json:"role_name"`
		JoinedAt  time.Time          `json:"joined_at"`
	}

	result := []MemberResponse{}
	for _, m := range memberships {
		var user models.User
		if err := database.GetCollection("users").FindOne(ctx, bson.M{"_id": m.UserID}).Decode(&user); err != nil {
			continue
		}
		result = append(result, MemberResponse{
			ID:        m.ID,
			UserID:    m.UserID,
			Name:      user.Name,
			Email:     user.Email,
			RoleFlags: m.RoleFlags,
			RoleName:  roleName(m.RoleFlags),
			JoinedAt:  m.JoinedAt,
		})
	}

	return c.JSON(result)
}

func InviteTeamMember(c *fiber.Ctx) error {
	inviterID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	teamID, err := primitive.ObjectIDFromHex(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	var body struct {
		Email     string `json:"email"`
		RoleFlags int    `json:"role_flags"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getTeamRole(ctx, teamID, inviterID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var invitee models.User
	if err := database.GetCollection("users").FindOne(ctx, bson.M{"email": body.Email}).Decode(&invitee); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User with that email not found"})
	}

	existing := database.GetCollection("team_members").FindOne(ctx, bson.M{"team_id": teamID, "user_id": invitee.ID})
	if existing.Err() == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User is already a member"})
	}

	flags := body.RoleFlags
	if flags == 0 {
		flags = RoleViewer
	}

	member := &models.TeamMember{
		ID:        primitive.NewObjectID(),
		TeamID:    teamID,
		UserID:    invitee.ID,
		RoleFlags: flags,
		InvitedBy: inviterID,
		JoinedAt:  time.Now(),
	}
	if _, err := database.GetCollection("team_members").InsertOne(ctx, member); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add member"})
	}

	var team models.Team
	if err := database.GetCollection("teams").FindOne(ctx, bson.M{"_id": teamID}).Decode(&team); err == nil {
		createNotification(ctx, invitee.ID, "team_invite",
			"You have been invited to team \""+team.Name+"\"",
			primitive.NilObjectID, primitive.NilObjectID)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Member added successfully",
		"member": fiber.Map{
			"user_id":    invitee.ID,
			"name":       invitee.Name,
			"email":      invitee.Email,
			"role_flags": flags,
			"role_name":  roleName(flags),
		},
	})
}

func UpdateTeamMemberRole(c *fiber.Ctx) error {
	requesterID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	teamID, err := primitive.ObjectIDFromHex(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	targetUserID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var body struct {
		RoleFlags int `json:"role_flags"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getTeamRole(ctx, teamID, requesterID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	if _, err := database.GetCollection("team_members").UpdateOne(ctx,
		bson.M{"team_id": teamID, "user_id": targetUserID},
		bson.M{"$set": bson.M{"role_flags": body.RoleFlags}},
	); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update role"})
	}

	return c.JSON(fiber.Map{
		"user_id":    targetUserID,
		"role_flags": body.RoleFlags,
		"role_name":  roleName(body.RoleFlags),
	})
}

func RemoveTeamMember(c *fiber.Ctx) error {
	requesterID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	teamID, err := primitive.ObjectIDFromHex(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	targetUserID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getTeamRole(ctx, teamID, requesterID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("team_members").DeleteOne(ctx, bson.M{"team_id": teamID, "user_id": targetUserID})
	return c.JSON(fiber.Map{"message": "Member removed"})
}

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

func uploadTeamImage(c *fiber.Ctx, imageType string) error {
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
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No file provided"})
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedImageExts[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid image type"})
	}

	dir := filepath.Join("../data/teams", teamID.Hex())
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("uploadTeamImage MkdirAll error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create storage directory"})
	}

	existingGlob := filepath.Join(dir, imageType+".*")
	if matches, _ := filepath.Glob(existingGlob); len(matches) > 0 {
		for _, m := range matches {
			os.Remove(m)
		}
	}

	destPath := filepath.Join(dir, imageType+ext)
	if err := c.SaveFile(fh, destPath); err != nil {
		log.Printf("uploadTeamImage SaveFile error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save image"})
	}

	imageURL := fmt.Sprintf("/api/team-media/%s/%s", teamID.Hex(), imageType)
	field := imageType + "_url"

	col := database.GetCollection("teams")
	if _, err := col.UpdateOne(ctx, bson.M{"_id": teamID}, bson.M{"$set": bson.M{
		field:        imageURL,
		"updated_at": time.Now(),
	}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update team"})
	}

	var team models.Team
	col.FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)

	count, _ := database.GetCollection("team_members").CountDocuments(ctx, bson.M{"team_id": teamID})

	return c.JSON(fiber.Map{
		"id":           team.ID,
		"name":         team.Name,
		"workspace_id": team.WorkspaceID,
		"avatar_url":   team.AvatarURL,
		"banner_url":   team.BannerURL,
		"member_count": count,
		"role_flags":   roleFlags,
		"role_name":    roleName(roleFlags),
		"created_at":   team.CreatedAt,
		"updated_at":   team.UpdatedAt,
	})
}

func UploadTeamAvatar(c *fiber.Ctx) error {
	return uploadTeamImage(c, "avatar")
}

func UploadTeamBanner(c *fiber.Ctx) error {
	return uploadTeamImage(c, "banner")
}

func serveTeamImage(c *fiber.Ctx, imageType string) error {
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

	dir := filepath.Join("../data/teams", teamID.Hex())
	for ext := range allowedImageExts {
		p := filepath.Join(dir, imageType+ext)
		if _, err := os.Stat(p); err == nil {
			return c.SendFile(p)
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Image not found"})
}

func ServeTeamAvatar(c *fiber.Ctx) error {
	return serveTeamImage(c, "avatar")
}

func ServeTeamBanner(c *fiber.Ctx) error {
	return serveTeamImage(c, "banner")
}

func ServePublicTeamImage(c *fiber.Ctx) error {
	teamID := c.Params("teamId")
	imageType := c.Params("imageType")
	if _, err := primitive.ObjectIDFromHex(teamID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}
	if imageType != "avatar" && imageType != "banner" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid image type"})
	}

	dir := filepath.Join("../data/teams", teamID)
	for ext := range allowedImageExts {
		p := filepath.Join(dir, imageType+ext)
		if _, err := os.Stat(p); err == nil {
			c.Set("Cache-Control", "public, max-age=3600")
			return c.SendFile(p)
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Image not found"})
}
