package models

import (
	"time"

	"github.com/google/uuid"
)

// type Reminder struct {
// 	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
// 	UserID    uuid.UUID
// 	MessageID uuid.UUID
// 	RemindAt  time.Time
// 	Notified  bool
// }

type Reminder struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID  // Who the reminder is for
	MessageID *uuid.UUID // (Optional but powerful) Links to a message you're reminding about
	Title     string     // A short title or description for the reminder
	Time      time.Time  // Original time (can be used for reference)
	RemindAt  time.Time  // ⏰ When the reminder should trigger
	CreatedAt time.Time  // When it was created
	Notified  bool       `gorm:"default:false"` // ✅ Tells us if reminder has been sent already
}
