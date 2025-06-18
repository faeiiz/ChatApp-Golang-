package main

import (
	"chatAppDemo/db"
	"chatAppDemo/handlers"
	"chatAppDemo/models"
	"chatAppDemo/routes"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func main() {
	app := fiber.New()

	// Connect DB
	db.ConnectToDatabase()

	// Auto migrate
	db.DB.AutoMigrate(
		&models.UserDemo{},
		&models.Message{},
		&models.Reminder{},
		&models.Notification{},
		&models.Group{},
		&models.GroupMember{},
		&models.GroupMessageDelivery{},
	)

	// Routes
	routes.SetupRoutes(app)

	go func() {
		for {
			handlers.CheckReminders()    // Call your checker
			time.Sleep(30 * time.Second) // Wait between checks
			fmt.Println("Checking reminders at", time.Now().Format(time.RFC822))

		}
	}()
	// CreateTestGroup()

	app.Listen(":3000")
}
func CreateTestGroup() {
	group := models.Group{Name: "Test Group"}
	db.DB.Create(&group)

	// Add 3 users to the group (replace with real user IDs)
	memberIDs := []uuid.UUID{
		uuid.MustParse("7b9ef093-2f07-4aa9-832a-56f3fdedaa6b"),
		uuid.MustParse("ccc7213f-22a8-4998-9476-f3c346ae0462"),
		uuid.MustParse("d4103b76-aa86-4b11-8ee9-f7c47d7645e4"),
	}

	for _, uid := range memberIDs {
		member := models.GroupMember{
			GroupID: group.ID,
			UserID:  uid,
		}
		db.DB.Create(&member)
	}

	fmt.Println("Group created with ID:", group.ID)
}
