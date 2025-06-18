package models

import (
	"time"

	"github.com/google/uuid"
)

type UserDemo struct {
	ID    uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name  string
	Email string
}
type Message struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	SenderID   uuid.UUID
	ReceiverID *uuid.UUID // nullable for group
	GroupID    *uuid.UUID // nullable for personal chat
	Content    string
	CreatedAt  time.Time
	ReadAt     *time.Time
	Delivered  bool `gorm:"default:false"` // <-- add this
}
type MessagePayload struct {
	ID        uuid.UUID `json:"id"`
	From      uuid.UUID `json:"from"`
	To        uuid.UUID `json:"to"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
