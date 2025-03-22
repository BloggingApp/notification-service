package model

import "github.com/google/uuid"

type Follower struct {
	UserID                      uuid.UUID `json:"user_id"`
	FollowerID                  uuid.UUID `json:"follower_id"`
	NewPostNotificationsEnabled bool      `json:"new_post_notifications_enabled"`
}
