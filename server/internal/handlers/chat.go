package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/fpmb/server/internal/database"
	"github.com/fpmb/server/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type chatRoom struct {
	clients map[*websocket.Conn]*wsClient
	mu      sync.RWMutex
}

var chatRooms = struct {
	m  map[string]*chatRoom
	mu sync.RWMutex
}{m: make(map[string]*chatRoom)}

func getChatRoom(teamID string) *chatRoom {
	chatRooms.mu.Lock()
	defer chatRooms.mu.Unlock()
	if room, ok := chatRooms.m[teamID]; ok {
		return room
	}
	room := &chatRoom{clients: make(map[*websocket.Conn]*wsClient)}
	chatRooms.m[teamID] = room
	return room
}

func (r *chatRoom) broadcast(sender *websocket.Conn, msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for conn := range r.clients {
		if conn != sender {
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func (r *chatRoom) broadcastAll(msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for conn := range r.clients {
		_ = conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (r *chatRoom) onlineUsers() []map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := map[string]bool{}
	list := make([]map[string]string, 0)
	for _, c := range r.clients {
		if !seen[c.userID] {
			seen[c.userID] = true
			list = append(list, map[string]string{"user_id": c.userID, "name": c.name})
		}
	}
	return list
}

func ListChatMessages(c *fiber.Ctx) error {
	teamID := c.Params("teamId")
	teamOID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid team ID"})
	}

	limitStr := c.Query("limit", "50")
	limit := int64(50)
	if l, err := primitive.ParseDecimal128(limitStr); err == nil {
		if s := l.String(); s != "" {
			if n, err := parseIntFromString(s); err == nil && n > 0 && n <= 200 {
				limit = n
			}
		}
	}

	beforeStr := c.Query("before", "")
	filter := bson.M{"team_id": teamOID}
	if beforeStr != "" {
		if beforeID, err := primitive.ObjectIDFromHex(beforeStr); err == nil {
			filter["_id"] = bson.M{"$lt": beforeID}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(limit)
	cursor, err := database.GetCollection("chat_messages").Find(ctx, filter, opts)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch messages"})
	}
	defer cursor.Close(ctx)

	var messages []models.ChatMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode messages"})
	}
	if messages == nil {
		messages = []models.ChatMessage{}
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return c.JSON(messages)
}

func parseIntFromString(s string) (int64, error) {
	var n int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}

func TeamChatWS(c *websocket.Conn) {
	teamID := c.Params("id")
	tokenStr := c.Query("token", "")
	userName := c.Query("name", "Anonymous")

	userID, _, ok := parseWSToken(tokenStr)
	if !ok {
		_ = c.WriteJSON(map[string]string{"type": "error", "message": "unauthorized"})
		_ = c.Close()
		return
	}

	room := getChatRoom(teamID)

	client := &wsClient{conn: c, userID: userID, name: userName}
	room.mu.Lock()
	room.clients[c] = client
	room.mu.Unlock()

	presenceMsg, _ := json.Marshal(map[string]interface{}{
		"type":  "presence",
		"users": room.onlineUsers(),
	})
	room.broadcastAll(presenceMsg)

	defer func() {
		room.mu.Lock()
		delete(room.clients, c)
		empty := len(room.clients) == 0
		room.mu.Unlock()

		leaveMsg, _ := json.Marshal(map[string]interface{}{
			"type":  "presence",
			"users": room.onlineUsers(),
		})
		room.broadcast(nil, leaveMsg)

		if empty {
			chatRooms.mu.Lock()
			delete(chatRooms.m, teamID)
			chatRooms.mu.Unlock()
		}
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				// unexpected error
			}
			break
		}

		var incoming struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if json.Unmarshal(msg, &incoming) != nil {
			continue
		}

		if incoming.Type == "message" {
			content := strings.TrimSpace(incoming.Content)
			if content == "" || len(content) > 5000 {
				continue
			}

			teamOID, err := primitive.ObjectIDFromHex(teamID)
			if err != nil {
				continue
			}
			userOID, err := primitive.ObjectIDFromHex(userID)
			if err != nil {
				continue
			}

			chatMsg := models.ChatMessage{
				ID:        primitive.NewObjectID(),
				TeamID:    teamOID,
				UserID:    userOID,
				UserName:  userName,
				Content:   content,
				CreatedAt: time.Now(),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, _ = database.GetCollection("chat_messages").InsertOne(ctx, chatMsg)
			cancel()

			outMsg, _ := json.Marshal(map[string]interface{}{
				"type":    "message",
				"message": chatMsg,
			})
			room.broadcastAll(outMsg)
		}

		if incoming.Type == "typing" {
			typingMsg, _ := json.Marshal(map[string]interface{}{
				"type":    "typing",
				"user_id": userID,
				"name":    userName,
			})
			room.broadcast(c, typingMsg)
		}
	}
}
