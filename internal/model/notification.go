package model

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}
