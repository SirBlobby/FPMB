package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	Name         string             `bson:"name"                 json:"name"`
	Email        string             `bson:"email"                json:"email"`
	PasswordHash string             `bson:"password_hash"        json:"-"`
	AvatarURL    string             `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	CreatedAt    time.Time          `bson:"created_at"           json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"           json:"updated_at"`
}

type Team struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"          json:"id"`
	Name        string             `bson:"name"                   json:"name"`
	WorkspaceID string             `bson:"workspace_id"           json:"workspace_id"`
	AvatarURL   string             `bson:"avatar_url,omitempty"   json:"avatar_url,omitempty"`
	BannerURL   string             `bson:"banner_url,omitempty"   json:"banner_url,omitempty"`
	CreatedBy   primitive.ObjectID `bson:"created_by"             json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at"             json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"             json:"updated_at"`
}

type TeamMember struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TeamID    primitive.ObjectID `bson:"team_id"       json:"team_id"`
	UserID    primitive.ObjectID `bson:"user_id"       json:"user_id"`
	RoleFlags int                `bson:"role_flags"    json:"role_flags"`
	InvitedBy primitive.ObjectID `bson:"invited_by"    json:"invited_by"`
	JoinedAt  time.Time          `bson:"joined_at"     json:"joined_at"`
}

type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	TeamID      primitive.ObjectID `bson:"team_id"              json:"team_id"`
	Name        string             `bson:"name"                 json:"name"`
	Description string             `bson:"description"          json:"description"`
	Visibility  string             `bson:"visibility"           json:"visibility"`
	IsPublic    bool               `bson:"is_public"            json:"is_public"`
	IsArchived  bool               `bson:"is_archived"          json:"is_archived"`
	CreatedBy   primitive.ObjectID `bson:"created_by"           json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at"           json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"           json:"updated_at"`
}

type ProjectMember struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id"    json:"project_id"`
	UserID    primitive.ObjectID `bson:"user_id"       json:"user_id"`
	RoleFlags int                `bson:"role_flags"    json:"role_flags"`
	AddedAt   time.Time          `bson:"added_at"      json:"added_at"`
}

type BoardColumn struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id"    json:"project_id"`
	Title     string             `bson:"title"         json:"title"`
	Position  int                `bson:"position"      json:"position"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"    json:"updated_at"`
}

type Subtask struct {
	ID   int    `bson:"id"   json:"id"`
	Text string `bson:"text" json:"text"`
	Done bool   `bson:"done" json:"done"`
}

type Card struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	ColumnID         primitive.ObjectID `bson:"column_id"            json:"column_id"`
	ProjectID        primitive.ObjectID `bson:"project_id"           json:"project_id"`
	Title            string             `bson:"title"                json:"title"`
	Description      string             `bson:"description"          json:"description"`
	Priority         string             `bson:"priority"             json:"priority"`
	Color            string             `bson:"color"                json:"color"`
	DueDate          *time.Time         `bson:"due_date,omitempty"   json:"due_date,omitempty"`
	Assignees        []string           `bson:"assignees"            json:"assignees"`
	EstimatedMinutes *int               `bson:"estimated_minutes,omitempty" json:"estimated_minutes,omitempty"`
	ActualMinutes    *int               `bson:"actual_minutes,omitempty"    json:"actual_minutes,omitempty"`
	Subtasks         []Subtask          `bson:"subtasks"             json:"subtasks"`
	Position         int                `bson:"position"             json:"position"`
	CreatedBy        primitive.ObjectID `bson:"created_by"           json:"created_by"`
	CreatedAt        time.Time          `bson:"created_at"           json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at"           json:"updated_at"`
}

type Event struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	Title       string             `bson:"title"                json:"title"`
	Date        string             `bson:"date"                 json:"date"`
	Time        string             `bson:"time"                 json:"time"`
	Color       string             `bson:"color"                json:"color"`
	Description string             `bson:"description"          json:"description"`
	Scope       string             `bson:"scope"                json:"scope"`
	ScopeID     primitive.ObjectID `bson:"scope_id"             json:"scope_id"`
	CreatedBy   primitive.ObjectID `bson:"created_by"           json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at"           json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"           json:"updated_at"`
}

type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	UserID    primitive.ObjectID `bson:"user_id"              json:"user_id"`
	Type      string             `bson:"type"                 json:"type"`
	Message   string             `bson:"message"              json:"message"`
	ProjectID primitive.ObjectID `bson:"project_id"           json:"project_id"`
	CardID    primitive.ObjectID `bson:"card_id,omitempty"    json:"card_id,omitempty"`
	Read      bool               `bson:"read"                 json:"read"`
	CreatedAt time.Time          `bson:"created_at"           json:"created_at"`
}

type Doc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TeamID    primitive.ObjectID `bson:"team_id"       json:"team_id"`
	Title     string             `bson:"title"         json:"title"`
	Content   string             `bson:"content"       json:"content"`
	CreatedBy primitive.ObjectID `bson:"created_by"    json:"created_by"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"    json:"updated_at"`
}

type File struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty"          json:"id"`
	ProjectID  primitive.ObjectID  `bson:"project_id,omitempty"   json:"project_id,omitempty"`
	TeamID     primitive.ObjectID  `bson:"team_id,omitempty"      json:"team_id,omitempty"`
	UserID     primitive.ObjectID  `bson:"user_id,omitempty"      json:"user_id,omitempty"`
	Name       string              `bson:"name"                   json:"name"`
	Type       string              `bson:"type"                   json:"type"`
	SizeBytes  int64               `bson:"size_bytes"             json:"size_bytes"`
	ParentID   *primitive.ObjectID `bson:"parent_id,omitempty"    json:"parent_id,omitempty"`
	StorageURL string              `bson:"storage_url,omitempty"  json:"storage_url,omitempty"`
	CreatedBy  primitive.ObjectID  `bson:"created_by"             json:"created_by"`
	CreatedAt  time.Time           `bson:"created_at"             json:"created_at"`
	UpdatedAt  time.Time           `bson:"updated_at"             json:"updated_at"`
}

type Webhook struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"            json:"id"`
	ProjectID     primitive.ObjectID `bson:"project_id"               json:"project_id"`
	Name          string             `bson:"name"                     json:"name"`
	Type          string             `bson:"type"                     json:"type"`
	URL           string             `bson:"url"                      json:"url"`
	SecretHash    string             `bson:"secret_hash,omitempty"    json:"-"`
	Status        string             `bson:"status"                   json:"status"`
	LastTriggered *time.Time         `bson:"last_triggered,omitempty" json:"last_triggered,omitempty"`
	CreatedBy     primitive.ObjectID `bson:"created_by"               json:"created_by"`
	CreatedAt     time.Time          `bson:"created_at"               json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"               json:"updated_at"`
}

type Whiteboard struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id"    json:"project_id"`
	Data      string             `bson:"data"          json:"data"`
	CreatedBy primitive.ObjectID `bson:"created_by"    json:"created_by"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"    json:"updated_at"`
}

type APIKey struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"          json:"id"`
	UserID    primitive.ObjectID `bson:"user_id"                json:"user_id"`
	Name      string             `bson:"name"                   json:"name"`
	Scopes    []string           `bson:"scopes"                 json:"scopes"`
	KeyHash   string             `bson:"key_hash"               json:"-"`
	Prefix    string             `bson:"prefix"                 json:"prefix"`
	LastUsed  *time.Time         `bson:"last_used,omitempty"    json:"last_used,omitempty"`
	RevokedAt *time.Time         `bson:"revoked_at,omitempty"   json:"revoked_at,omitempty"`
	CreatedAt time.Time          `bson:"created_at"             json:"created_at"`
}

type ChatMessage struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty"          json:"id"`
	TeamID    primitive.ObjectID  `bson:"team_id"                json:"team_id"`
	UserID    primitive.ObjectID  `bson:"user_id"                json:"user_id"`
	UserName  string              `bson:"user_name"              json:"user_name"`
	Content   string              `bson:"content"                json:"content"`
	ReplyTo   *primitive.ObjectID `bson:"reply_to,omitempty"     json:"reply_to,omitempty"`
	EditedAt  *time.Time          `bson:"edited_at,omitempty"    json:"edited_at,omitempty"`
	Deleted   bool                `bson:"deleted,omitempty"      json:"deleted,omitempty"`
	CreatedAt time.Time           `bson:"created_at"             json:"created_at"`
}
