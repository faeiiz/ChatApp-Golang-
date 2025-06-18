package routes

import (
	"chatAppDemo/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/messages/send", handlers.SendMessage)
	api.Post("/reminders", handlers.SetReminder)
	api.Get("/notifications", handlers.GetNotifications)
	api.Get("/reminders/check", TriggerReminderCheck)

	app.Use("/ws", handlers.WebSocketHandler)                   // Check & store user ID from query
	app.Get("/ws", websocket.New(handlers.WebSocketConnection)) // Start actual WS

}
func TriggerReminderCheck(c *fiber.Ctx) error {
	handlers.CheckReminders() // this is your function
	return c.JSON(fiber.Map{"status": "Checked reminders"})
}
