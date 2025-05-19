package model

import (
	"time"

	"github.com/google/uuid"
)

type GlobalNotification struct {
	ID           int64     `json:"id"`
	PosterID     uuid.UUID `json:"poster_id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	ResourceLink string    `json:"resource_link"`
	CreatedAt    time.Time `json:"created_at"`
}
