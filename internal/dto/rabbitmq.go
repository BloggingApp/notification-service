package dto

import (
	"time"

	"github.com/google/uuid"
)

type MQNotificateUserCode struct {
	Email string `json:"email"`
	Code  int    `json:"code"`
}

type MQUserCreated struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName *string   `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
}

type MQPostCreated struct {
	PostID    int64     `json:"post_id"`
	UserID    uuid.UUID `json:"user_id"`
	PostTitle string    `json:"post_title"`
	CreatedAt time.Time `json:"created_at"`
}

type MQFollow struct {
	UserID     uuid.UUID `json:"user_id"`
	FollowerID uuid.UUID `json:"follower_id"`
}

type MQNewPostNotificationsEnabledUpdate struct {
	UserID     uuid.UUID `json:"user_id"`
	FollowerID uuid.UUID `json:"follower_id"`
	Enabled    bool      `json:"enabled"`
}

type MQPostValidationStatusUpdate struct {
	PostID    int64      `json:"post_id"`
	UserID    uuid.UUID  `json:"user_id"`
	StatusMsg string     `json:"status_msg"`
}
