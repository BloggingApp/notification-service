package model

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Content    string    `json:"content"`
	ResourceID string    `json:"resource_id"`
	CreatedAt  time.Time `json:"created_at"`
}
