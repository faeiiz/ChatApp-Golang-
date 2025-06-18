package models

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name      string
	CreatedAt time.Time
}
type GroupMember struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	GroupID uuid.UUID
	UserID  uuid.UUID
}
type GroupMessageDelivery struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	MessageID uuid.UUID
	UserID    uuid.UUID
	Delivered bool
	CreatedAt time.Time
}
