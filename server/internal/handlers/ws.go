package handlers

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
)

type wsMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	UserID  string          `json:"user_id,omitempty"`
	Name    string          `json:"name,omitempty"`
	X       float64         `json:"x,omitempty"`
	Y       float64         `json:"y,omitempty"`
}

type wsClient struct {
	conn   *websocket.Conn
	userID string
	name   string
}

type whiteboardRoom struct {
	clients map[*websocket.Conn]*wsClient
	mu      sync.RWMutex
}

var wsRooms = struct {
	m  map[string]*whiteboardRoom
	mu sync.RWMutex
}{m: make(map[string]*whiteboardRoom)}

func getRoom(boardID string) *whiteboardRoom {
	wsRooms.mu.Lock()
	defer wsRooms.mu.Unlock()
	if room, ok := wsRooms.m[boardID]; ok {
		return room
	}
	room := &whiteboardRoom{clients: make(map[*websocket.Conn]*wsClient)}
	wsRooms.m[boardID] = room
	return room
}

func (r *whiteboardRoom) broadcast(sender *websocket.Conn, msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for conn := range r.clients {
		if conn != sender {
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func (r *whiteboardRoom) userList() []map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]map[string]string, 0, len(r.clients))
	for _, c := range r.clients {
		list = append(list, map[string]string{"user_id": c.userID, "name": c.name})
	}
	return list
}

func parseWSToken(tokenStr string) (userID string, email string, ok bool) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "changeme-jwt-secret"
	}

	type claims struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		jwt.RegisteredClaims
	}
	c := &claims{}
	token, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", false
	}
	return c.UserID, c.Email, true
}

func WhiteboardWS(c *websocket.Conn) {
	boardID := c.Params("id")
	tokenStr := c.Query("token", "")
	userName := c.Query("name", "Anonymous")

	userID, _, ok := parseWSToken(tokenStr)
	if !ok {
		_ = c.WriteJSON(map[string]string{"type": "error", "payload": "unauthorized"})
		_ = c.Close()
		return
	}

	room := getRoom(boardID)

	client := &wsClient{conn: c, userID: userID, name: userName}
	room.mu.Lock()
	room.clients[c] = client
	room.mu.Unlock()

	joinMsg, _ := json.Marshal(map[string]interface{}{
		"type":    "join",
		"user_id": userID,
		"name":    userName,
		"users":   room.userList(),
	})
	room.broadcast(nil, joinMsg)

	selfMsg, _ := json.Marshal(map[string]interface{}{
		"type":  "users",
		"users": room.userList(),
	})
	_ = c.WriteMessage(websocket.TextMessage, selfMsg)

	defer func() {
		room.mu.Lock()
		delete(room.clients, c)
		empty := len(room.clients) == 0
		room.mu.Unlock()

		leaveMsg, _ := json.Marshal(map[string]interface{}{
			"type":    "leave",
			"user_id": userID,
			"name":    userName,
			"users":   room.userList(),
		})
		room.broadcast(nil, leaveMsg)

		if empty {
			wsRooms.mu.Lock()
			delete(wsRooms.m, boardID)
			wsRooms.mu.Unlock()
		}
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				log.Printf("WS error board=%s user=%s: %v", boardID, userID, err)
			}
			break
		}

		var incoming map[string]interface{}
		if json.Unmarshal(msg, &incoming) == nil {
			incoming["user_id"] = userID
			incoming["name"] = userName
			outMsg, _ := json.Marshal(incoming)
			room.broadcast(c, outMsg)
		}
	}
}
