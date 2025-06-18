package handlers

import (
	"chatAppDemo/db"
	"chatAppDemo/models"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type MessageInput struct {
	SenderID   uuid.UUID `json:"sender_id"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Content    string    `json:"content"`
}

func SendMessage(c *fiber.Ctx) error {
	var input MessageInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	message := models.Message{
		SenderID:   input.SenderID,
		ReceiverID: &input.ReceiverID,
		Content:    input.Content,
		CreatedAt:  time.Now(),
	}
	db.DB.Create(&message)

	notif := models.Notification{
		UserID:    input.ReceiverID,
		Message:   "New message from user",
		Type:      "chat",
		Seen:      false,
		CreatedAt: time.Now(),
	}
	db.DB.Create(&notif)

	payload, _ := json.Marshal(fiber.Map{
		"type":    "chat",
		"message": input.Content,
		"from":    input.SenderID.String(),
	})

	SendToUser(input.ReceiverID, payload)

	return c.JSON(fiber.Map{"message": "Message sent"})
}
func HandleMessage(c *websocket.Conn) {
	for {
		var msg models.MessagePayload
		if err := c.ReadJSON(&msg); err != nil {
			log.Println("Error reading JSON:", err)
			break
		}

		fmt.Println("Message received:", msg)

		msg.ID = uuid.New()
		msg.Timestamp = time.Now()

		// Save message to DB
		if err := db.DB.Create(&msg).Error; err != nil {
			log.Println("Failed to save message:", err)
			continue
		}

		// Try to send to recipient if connected
		recipientConn, ok := Connections[msg.To]
		if ok {
			err := recipientConn.WriteJSON(msg)
			if err != nil {
				log.Println("Error sending to recipient:", err)
				notif := models.Notification{
					UserID:    msg.To,
					Message:   "New message from " + msg.From.String(),
					Type:      "chat",
					Seen:      false,
					CreatedAt: time.Now(),
				}
				db.DB.Create(&notif)

			} else {
				fmt.Println("Message sent to:", msg.To)
			}
		} else {
			fmt.Println("User", msg.To, "is not connected — maybe save reminder here")
		}
	}
}
