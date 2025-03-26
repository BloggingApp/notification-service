package service

import (
	"context"

	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/BloggingApp/notification-service/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type User interface {
	create(ctx context.Context, user model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	updateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	StartCreating(ctx context.Context)
	StartUpdating(ctx context.Context)
	StartCreatingFollowers(ctx context.Context)
}

type Notification interface {
	StartProcessingNewPostNotifications(ctx context.Context)
	GetUserNotifications(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*model.Notification, error)
	StartJobs()
}

type Service struct {
	User
	Notification
}

func New(logger *zap.Logger, repo *repository.Repository, rdb *redis.Client, rabbitmq *rabbitmq.MQConn) *Service {
	return &Service{
		User: newUserService(logger, repo, rdb, rabbitmq),
		Notification: newNotificationService(logger, repo, rdb, rabbitmq),
	}
}
