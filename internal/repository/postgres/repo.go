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
	UpdateFollowerNewPostNotifications(ctx context.Context, follower model.Follower) error
}

type Notification interface {
	GetInterestedFollowers(ctx context.Context, authorID uuid.UUID) ([]uuid.UUID, error)
	CreateBatch(ctx context.Context, notifications []model.Notification) error
	CreateBatched(ctx context.Context, notifications []model.Notification, batchSize int) error
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
