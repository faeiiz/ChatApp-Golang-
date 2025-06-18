package handlers

import (
	"chatAppDemo/db"
	"chatAppDemo/models"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func SetReminder(c *fiber.Ctx) error {
	type ReminderInput struct {
		UserID uuid.UUID `json:"user_id"` // Who to notify
		Title  string    `json:"title"`
		Time   time.Time `json:"time"`
	}

	var input ReminderInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	reminder := models.Reminder{
		UserID: input.UserID,
		Title:  input.Title,

		Time:      input.Time,
		CreatedAt: time.Now(),
	}
	db.DB.Create(&reminder)

	// 🟡 Send WebSocket notification if user is connected
	payload, _ := json.Marshal(fiber.Map{
		"type":  "reminder",
		"title": input.Title,
		"time":  input.Time,
	})

	SendToUser(input.UserID, payload)

	return c.JSON(fiber.Map{"message": "Reminder set"})
}

func CheckReminders() {
	var reminders []models.Reminder
	now := time.Now()

	db.DB.Where("remind_at <= ? AND notified = false", now).Find(&reminders)

	for _, r := range reminders {
		var msg models.Message
		db.DB.First(&msg, "id = ?", r.MessageID)

		notif := models.Notification{
			UserID:    r.UserID,
			Message:   "Reminder: " + msg.Content,
			Type:      "reminder",
			Seen:      false,
			CreatedAt: now,
		}
		db.DB.Create(&notif)
		// Send via WebSocket if user is online
		payload, _ := json.Marshal(fiber.Map{
			"type":    "reminder",
			"title":   r.Title,
			"time":    r.RemindAt,
			"message": notif.Message,
		})
		SendToUser(r.UserID, payload)

		r.Notified = true
		db.DB.Save(&r)
	}
}
