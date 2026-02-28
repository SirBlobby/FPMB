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

func getProjectRole(ctx context.Context, projectID, userID primitive.ObjectID) (int, error) {
	var pm models.ProjectMember
	err := database.GetCollection("project_members").FindOne(ctx, bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}).Decode(&pm)
	if err == nil {
		return pm.RoleFlags, nil
	}

	var project models.Project
	if err := database.GetCollection("projects").FindOne(ctx, bson.M{"_id": projectID}).Decode(&project); err != nil {
		return 0, err
	}

	if project.TeamID == primitive.NilObjectID {
		return 0, fiber.ErrForbidden
	}

	return getTeamRole(ctx, project.TeamID, userID)
}

func ListProjects(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type ProjectResponse struct {
		ID          primitive.ObjectID `json:"id"`
		Name        string             `json:"name"`
		Description string             `json:"description"`
		TeamID      primitive.ObjectID `json:"team_id"`
		TeamName    string             `json:"team_name"`
		RoleFlags   int                `json:"role_flags"`
		RoleName    string             `json:"role_name"`
		IsPublic    bool               `json:"is_public"`
		IsArchived  bool               `json:"is_archived"`
		UpdatedAt   time.Time          `json:"updated_at"`
	}

	result := []ProjectResponse{}

	cursor, err := database.GetCollection("team_members").Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch teams"})
	}
	defer cursor.Close(ctx)

	var memberships []models.TeamMember
	cursor.All(ctx, &memberships)

	for _, m := range memberships {
		var team models.Team
		database.GetCollection("teams").FindOne(ctx, bson.M{"_id": m.TeamID}).Decode(&team)

		projCursor, err := database.GetCollection("projects").Find(ctx, bson.M{"team_id": m.TeamID})
		if err != nil {
			continue
		}
		var projects []models.Project
		projCursor.All(ctx, &projects)
		projCursor.Close(ctx)

		for _, p := range projects {
			roleFlags := m.RoleFlags
			var pm models.ProjectMember
			if err := database.GetCollection("project_members").FindOne(ctx, bson.M{
				"project_id": p.ID,
				"user_id":    userID,
			}).Decode(&pm); err == nil {
				roleFlags = pm.RoleFlags
			}

			result = append(result, ProjectResponse{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				TeamID:      p.TeamID,
				TeamName:    team.Name,
				RoleFlags:   roleFlags,
				RoleName:    roleName(roleFlags),
				IsPublic:    p.IsPublic,
				IsArchived:  p.IsArchived,
				UpdatedAt:   p.UpdatedAt,
			})
		}
	}

	personalCursor, err := database.GetCollection("project_members").Find(ctx, bson.M{
		"user_id": userID,
	})
	if err == nil {
		defer personalCursor.Close(ctx)
		var pms []models.ProjectMember
		personalCursor.All(ctx, &pms)
		for _, pm := range pms {
			var p models.Project
			if err := database.GetCollection("projects").FindOne(ctx, bson.M{
				"_id":     pm.ProjectID,
				"team_id": primitive.NilObjectID,
			}).Decode(&p); err != nil {
				continue
			}
			result = append(result, ProjectResponse{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				TeamID:      p.TeamID,
				TeamName:    "Personal",
				RoleFlags:   pm.RoleFlags,
				RoleName:    roleName(pm.RoleFlags),
				IsPublic:    p.IsPublic,
				IsArchived:  p.IsArchived,
				UpdatedAt:   p.UpdatedAt,
			})
		}
	}

	return c.JSON(result)
}

func CreatePersonalProject(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Project name is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	project := &models.Project{
		ID:          primitive.NewObjectID(),
		TeamID:      primitive.NilObjectID,
		Name:        body.Name,
		Description: body.Description,
		IsPublic:    body.IsPublic,
		IsArchived:  false,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if _, err := database.GetCollection("projects").InsertOne(ctx, project); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create project"})
	}

	member := &models.ProjectMember{
		ID:        primitive.NewObjectID(),
		ProjectID: project.ID,
		UserID:    userID,
		RoleFlags: RoleOwner,
		AddedAt:   now,
	}
	database.GetCollection("project_members").InsertOne(ctx, member)

	defaultColumns := []string{"To Do", "In Progress", "Done"}
	for i, title := range defaultColumns {
		col := &models.BoardColumn{
			ID:        primitive.NewObjectID(),
			ProjectID: project.ID,
			Title:     title,
			Position:  i,
			CreatedAt: now,
			UpdatedAt: now,
		}
		database.GetCollection("board_columns").InsertOne(ctx, col)
	}

	return c.Status(fiber.StatusCreated).JSON(project)
}

func ListTeamProjects(c *fiber.Ctx) error {
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

	teamRole, err := getTeamRole(ctx, teamID, userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	cursor, err := database.GetCollection("projects").Find(ctx, bson.M{"team_id": teamID},
		options.Find().SetSort(bson.M{"updated_at": -1}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch projects"})
	}
	defer cursor.Close(ctx)

	var projects []models.Project
	cursor.All(ctx, &projects)

	type ProjectResponse struct {
		models.Project
		RoleFlags int    `json:"role_flags"`
		RoleName  string `json:"role_name"`
	}

	result := []ProjectResponse{}
	for _, p := range projects {
		flags := teamRole
		var pm models.ProjectMember
		if err := database.GetCollection("project_members").FindOne(ctx, bson.M{
			"project_id": p.ID, "user_id": userID,
		}).Decode(&pm); err == nil {
			flags = pm.RoleFlags
		}
		result = append(result, ProjectResponse{Project: p, RoleFlags: flags, RoleName: roleName(flags)})
	}

	return c.JSON(result)
}

func CreateProject(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	teamID, err := primitive.ObjectIDFromHex(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Project name is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getTeamRole(ctx, teamID, userID)
	if err != nil || !hasPermission(roleFlags, RoleEditor) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	now := time.Now()
	project := &models.Project{
		ID:          primitive.NewObjectID(),
		TeamID:      teamID,
		Name:        body.Name,
		Description: body.Description,
		IsPublic:    body.IsPublic,
		IsArchived:  false,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if _, err := database.GetCollection("projects").InsertOne(ctx, project); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create project"})
	}

	defaultColumns := []string{"To Do", "In Progress", "Done"}
	for i, title := range defaultColumns {
		col := &models.BoardColumn{
			ID:        primitive.NewObjectID(),
			ProjectID: project.ID,
			Title:     title,
			Position:  i,
			CreatedAt: now,
			UpdatedAt: now,
		}
		database.GetCollection("board_columns").InsertOne(ctx, col)
	}

	return c.Status(fiber.StatusCreated).JSON(project)
}

func GetProject(c *fiber.Ctx) error {
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

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	var project models.Project
	if err := database.GetCollection("projects").FindOne(ctx, bson.M{"_id": projectID}).Decode(&project); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	return c.JSON(fiber.Map{
		"id":          project.ID,
		"team_id":     project.TeamID,
		"name":        project.Name,
		"description": project.Description,
		"visibility":  project.Visibility,
		"is_public":   project.IsPublic,
		"is_archived": project.IsArchived,
		"role_flags":  roleFlags,
		"role_name":   roleName(roleFlags),
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
	})
}

func UpdateProject(c *fiber.Ctx) error {
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

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    *bool  `json:"is_public"`
		Visibility  string `json:"visibility"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	update := bson.M{"updated_at": time.Now()}
	if body.Name != "" {
		update["name"] = body.Name
	}
	if body.Description != "" {
		update["description"] = body.Description
	}
	if body.IsPublic != nil {
		update["is_public"] = *body.IsPublic
	}
	if body.Visibility != "" {
		update["visibility"] = body.Visibility
		update["is_public"] = body.Visibility == "public"
	}

	col := database.GetCollection("projects")
	col.UpdateOne(ctx, bson.M{"_id": projectID}, bson.M{"$set": update})

	var project models.Project
	col.FindOne(ctx, bson.M{"_id": projectID}).Decode(&project)
	return c.JSON(project)
}

func ArchiveProject(c *fiber.Ctx) error {
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

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var project models.Project
	if err := database.GetCollection("projects").FindOne(ctx, bson.M{"_id": projectID}).Decode(&project); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	database.GetCollection("projects").UpdateOne(ctx, bson.M{"_id": projectID}, bson.M{"$set": bson.M{
		"is_archived": !project.IsArchived,
		"updated_at":  time.Now(),
	}})

	return c.JSON(fiber.Map{"is_archived": !project.IsArchived})
}

func DeleteProject(c *fiber.Ctx) error {
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

	roleFlags, err := getProjectRole(ctx, projectID, userID)
	if err != nil || !hasPermission(roleFlags, RoleOwner) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only owners can delete projects"})
	}

	database.GetCollection("projects").DeleteOne(ctx, bson.M{"_id": projectID})
	database.GetCollection("board_columns").DeleteMany(ctx, bson.M{"project_id": projectID})
	database.GetCollection("cards").DeleteMany(ctx, bson.M{"project_id": projectID})
	database.GetCollection("project_members").DeleteMany(ctx, bson.M{"project_id": projectID})
	database.GetCollection("events").DeleteMany(ctx, bson.M{"scope_id": projectID, "scope": "project"})
	database.GetCollection("files").DeleteMany(ctx, bson.M{"project_id": projectID})
	database.GetCollection("webhooks").DeleteMany(ctx, bson.M{"project_id": projectID})
	database.GetCollection("whiteboards").DeleteMany(ctx, bson.M{"project_id": projectID})

	return c.JSON(fiber.Map{"message": "Project deleted"})
}

func ListProjectMembers(c *fiber.Ctx) error {
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

	cursor, err := database.GetCollection("project_members").Find(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch members"})
	}
	defer cursor.Close(ctx)

	var members []models.ProjectMember
	cursor.All(ctx, &members)

	type MemberResponse struct {
		UserID    primitive.ObjectID `json:"user_id"`
		Name      string             `json:"name"`
		Email     string             `json:"email"`
		RoleFlags int                `json:"role_flags"`
		RoleName  string             `json:"role_name"`
	}

	result := []MemberResponse{}
	for _, m := range members {
		var user models.User
		if err := database.GetCollection("users").FindOne(ctx, bson.M{"_id": m.UserID}).Decode(&user); err != nil {
			continue
		}
		result = append(result, MemberResponse{
			UserID:    m.UserID,
			Name:      user.Name,
			Email:     user.Email,
			RoleFlags: m.RoleFlags,
			RoleName:  roleName(m.RoleFlags),
		})
	}
	return c.JSON(result)
}

func AddProjectMember(c *fiber.Ctx) error {
	requesterID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	var body struct {
		UserID    string `json:"user_id"`
		RoleFlags int    `json:"role_flags"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	targetUserID, err := primitive.ObjectIDFromHex(body.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user_id"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, requesterID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	flags := body.RoleFlags
	if flags == 0 {
		flags = RoleViewer
	}

	member := &models.ProjectMember{
		ID:        primitive.NewObjectID(),
		ProjectID: projectID,
		UserID:    targetUserID,
		RoleFlags: flags,
		AddedAt:   time.Now(),
	}
	database.GetCollection("project_members").InsertOne(ctx, member)
	return c.Status(fiber.StatusCreated).JSON(member)
}

func UpdateProjectMemberRole(c *fiber.Ctx) error {
	requesterID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	targetUserID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var body struct {
		RoleFlags int `json:"role_flags"`
	}
	c.BodyParser(&body)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, requesterID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("project_members").UpdateOne(ctx,
		bson.M{"project_id": projectID, "user_id": targetUserID},
		bson.M{"$set": bson.M{"role_flags": body.RoleFlags}},
	)
	return c.JSON(fiber.Map{"user_id": targetUserID, "role_flags": body.RoleFlags, "role_name": roleName(body.RoleFlags)})
}

func RemoveProjectMember(c *fiber.Ctx) error {
	requesterID, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	projectID, err := primitive.ObjectIDFromHex(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID"})
	}

	targetUserID, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleFlags, err := getProjectRole(ctx, projectID, requesterID)
	if err != nil || !hasPermission(roleFlags, RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	database.GetCollection("project_members").DeleteOne(ctx, bson.M{"project_id": projectID, "user_id": targetUserID})
	return c.JSON(fiber.Map{"message": "Member removed"})
}
