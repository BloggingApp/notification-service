package postgres

import (
	"context"

	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User interface {
	Create(ctx context.Context, user model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	CreateFollower(ctx context.Context, follower model.Follower) error
	UpdateFollowerNewPostNotificationsEnabled(ctx context.Context, follower model.Follower) error
}

type Notification interface {
	GetInterestedFollowers(ctx context.Context, authorID uuid.UUID) ([]uuid.UUID, error)
	Create(ctx context.Context, notification model.Notification) error
	CreateBatch(ctx context.Context, notifications []model.Notification) error
	CreateBatched(ctx context.Context, notifications []model.Notification, batchSize int) error
	GetUserNotifications(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*model.Notification, error)
	DeleteOldNotifications(ctx context.Context) error
	CreateGlobalNotification(ctx context.Context, gn model.GlobalNotification) error
	GetGlobalNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.GlobalNotification, error)
	MarkGlobalNotificationAsRead(ctx context.Context, userID uuid.UUID, notificationID int64) error
}

type PGRepo struct {
	User
	Notification
}

func New(db *pgxpool.Pool) *PGRepo {
	return &PGRepo{
		User: newUserRepo(db),
		Notification: newNotificationRepo(db),
	}
}
