package handlers

import (
	"chatAppDemo/db"
	"chatAppDemo/models"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

var Connections = make(map[uuid.UUID]*websocket.Conn)
var mu sync.Mutex

func WebSocketHandler(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("user_id", userID)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func WebSocketConnection(c *websocket.Conn) {
	userID := c.Locals("user_id").(uuid.UUID)
	RegisterClient(userID, c)
	defer UnregisterClient(userID)

	// ✅ Step 1: Send undelivered messages
	unread := []models.Message{}
	if err := db.DB.Where("receiver_id = ? AND delivered = ?", userID, false).Find(&unread).Error; err != nil {
		log.Println("Error fetching unread messages:", err)
	} else {
		for _, msg := range unread {
			outgoing := fiber.Map{
				"type":    "chat",
				"from":    msg.SenderID,
				"message": msg.Content,
				"time":    msg.CreatedAt,
			}
			payload, _ := json.Marshal(outgoing)
			c.WriteMessage(websocket.TextMessage, payload)
			// Mark as delivered
			db.DB.Model(&msg).Update("delivered", true)
		}
	}

	// 🟡 Step 1.5: Send unseen reminder notifications
	go func() {
		var notifications []models.Notification
		if err := db.DB.Where("user_id = ? AND seen = false AND type = ?", userID, "reminder").Find(&notifications).Error; err == nil {
			for _, notif := range notifications {
				payload, _ := json.Marshal(fiber.Map{
					"type":    "reminder",
					"message": notif.Message,
					"time":    notif.CreatedAt,
				})
				c.WriteMessage(websocket.TextMessage, payload)
				db.DB.Model(&notif).Update("seen", true)
			}
		}
	}()
	go func() {
		var pendingDeliveries []models.GroupMessageDelivery
		if err := db.DB.Where("user_id = ? AND delivered = false", userID).Find(&pendingDeliveries).Error; err == nil {
			for _, delivery := range pendingDeliveries {
				var msg models.Message
				if err := db.DB.First(&msg, "id = ?", delivery.MessageID).Error; err != nil {
					continue
				}

				payload, _ := json.Marshal(fiber.Map{
					"type":     "group",
					"group_id": msg.GroupID,
					"from":     msg.SenderID,
					"message":  msg.Content,
					"time":     msg.CreatedAt,
				})
				c.WriteMessage(websocket.TextMessage, payload)
				db.DB.Model(&delivery).Update("delivered", true)
			}
		}
	}()

	// ✅ Step 2: Listen for new messages from this user
	type IncomingMessage struct {
		To      *uuid.UUID `json:"to,omitempty"`       // for personal
		GroupID *uuid.UUID `json:"group_id,omitempty"` // for group
		Content string     `json:"content"`
	}

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		if incoming.GroupID != nil {
			// 🚀 GROUP MESSAGE
			groupMessage := models.Message{
				SenderID:  userID,
				GroupID:   incoming.GroupID,
				Content:   incoming.Content,
				CreatedAt: time.Now(),
			}

			if err := db.DB.Create(&groupMessage).Error; err != nil {
				log.Println("Failed to save group message:", err)
				continue
			}

			// 📨 Broadcast to all group members
			var members []models.GroupMember
			db.DB.Where("group_id = ?", *incoming.GroupID).Find(&members)

			for _, member := range members {
				if member.UserID == userID {
					continue // skip sender
				}
				// ✅ Create delivery record
				delivery := models.GroupMessageDelivery{
					MessageID: groupMessage.ID,
					UserID:    member.UserID,
					Delivered: false,
					CreatedAt: time.Now(),
				}
				if err := db.DB.Create(&delivery).Error; err != nil {
					log.Println("❌ Failed to create delivery:", err)
				} else {
					log.Println("✅ Created delivery for user", member.UserID)
				}

				outgoing := fiber.Map{
					"type":     "group",
					"group_id": incoming.GroupID,
					"from":     userID,
					"message":  incoming.Content,
					"time":     groupMessage.CreatedAt,
				}
				payload, _ := json.Marshal(outgoing)

				mu.Lock()
				recipientConn, ok := Connections[member.UserID]
				mu.Unlock()

				if ok {
					recipientConn.WriteMessage(websocket.TextMessage, payload)
				}
			}

		} else if incoming.To != nil {
			// 🧍 PERSONAL MESSAGE (existing logic)
			message := models.Message{
				SenderID:   userID,
				ReceiverID: incoming.To,
				Content:    incoming.Content,
				CreatedAt:  time.Now(),
				Delivered:  false,
			}

			if err := db.DB.Create(&message).Error; err != nil {
				log.Println("Failed to save message:", err)
			}

			outgoing := fiber.Map{
				"type":    "chat",
				"from":    userID,
				"message": incoming.Content,
				"time":    message.CreatedAt,
			}
			payload, _ := json.Marshal(outgoing)

			mu.Lock()
			recipientConn, ok := Connections[*incoming.To]
			mu.Unlock()

			if ok {
				if err := recipientConn.WriteMessage(websocket.TextMessage, payload); err == nil {
					db.DB.Model(&message).Update("delivered", true)
				}
			}
		}
	}
}
func RegisterClient(userID uuid.UUID, conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	Connections[userID] = conn
}

// UnregisterClient removes a user from the Connections map
func UnregisterClient(userID uuid.UUID) {
	mu.Lock()
	defer mu.Unlock()
	delete(Connections, userID)
}

// SendToUser sends a byte message to a specific user if they're connected
func SendToUser(userID uuid.UUID, message []byte) {
	mu.Lock()
	conn, ok := Connections[userID]
	mu.Unlock()

	if ok {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}
