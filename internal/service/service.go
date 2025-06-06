package service

import (
	"context"

	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/BloggingApp/notification-service/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
	StartUpdatingFollowersNewPostNotificationsEnabled(ctx context.Context)
}

type Notification interface {
	RegisterConnection(userID uuid.UUID, conn *websocket.Conn)
	UnregisterConnection(userID uuid.UUID)
	StartProcessingNewPostNotifications(ctx context.Context)
	GetUserNotifications(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*model.Notification, error)
	StartJobs()
	CreateGlobalNotification(ctx context.Context, gn model.GlobalNotification) error
	GetGlobalNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.GlobalNotification, error)
	MarkGlobalNotificationAsRead(ctx context.Context, userID uuid.UUID, notificationID int64) error
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
