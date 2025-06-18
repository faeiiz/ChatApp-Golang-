package handlers

import (
	"chatAppDemo/db"
	"chatAppDemo/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetNotifications(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(400).SendString("Invalid user_id")
	}

	var notifs []models.Notification
	if err := db.DB.Where("user_id = ? AND seen = false", userID).Find(&notifs).Error; err != nil {
		return c.Status(500).SendString("DB error")
	}

	return c.JSON(notifs)
}
