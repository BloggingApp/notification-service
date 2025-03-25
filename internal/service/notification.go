package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BloggingApp/notification-service/internal/dto"
	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/BloggingApp/notification-service/internal/repository"
	"github.com/BloggingApp/notification-service/internal/repository/redisrepo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type notificationService struct {
	logger *zap.Logger
	repo *repository.Repository
	rdb *redis.Client
	rabbitmq *rabbitmq.MQConn
}

func newNotificationService(logger *zap.Logger, repo *repository.Repository, rdb *redis.Client, rabbitmq *rabbitmq.MQConn) Notification {
	return &notificationService{
		logger: logger,
		repo: repo,
		rdb: rdb,
		rabbitmq: rabbitmq,
	}
}

func (s *notificationService) StartProcessingNewPostNotifications(ctx context.Context) {
	msgs, err := s.rabbitmq.Consume(rabbitmq.NEW_POST_QUEUE)
	if err != nil {
		panic(err)
	}

	for msg := range msgs {
		var postCreatedDto dto.MQPostCreated
		if err := json.Unmarshal(msg.Body, &postCreatedDto); err != nil {
			msg.Ack(false)
			continue
		}

		receivers, err := s.repo.Postgres.Notification.GetInterestedFollowers(ctx, postCreatedDto.UserID)
		if err != nil {
			s.logger.Sugar().Errorf("failed to get user(%s)'s interested followers: %s", postCreatedDto.UserID.String(), err.Error())
			msg.Ack(false)
			continue
		}

		author, err := s.repo.Postgres.User.FindByID(ctx, postCreatedDto.UserID)
		if err != nil {
			s.logger.Sugar().Errorf("failed to get post(%d) author(%s) from postgres: %s", postCreatedDto.PostID, postCreatedDto.UserID.String(), err.Error())
			msg.Ack(false)
			continue
		}

		notificationType := "newpost"
		message := fmt.Sprintf("%s has created new post: %s", author.Username, postCreatedDto.PostTitle)

		var notifications []model.Notification
		for _, receiver := range receivers {
			notifications = append(notifications, model.Notification{
				Type: notificationType,
				ReceiverID: receiver,
				Message: message,
			})
		}

		if err := s.repo.Postgres.Notification.CreateBatched(ctx, notifications, 200); err != nil {
			s.logger.Sugar().Errorf("failed to create batched notifications for post(%d): %s", postCreatedDto.PostID, err.Error())
			msg.Ack(false)
			continue
		}

		msg.Ack(false)
	}
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*model.Notification, error) {
	notificationsCache, err := redisrepo.Get[[]*model.Notification](s.rdb, ctx, redisrepo.UserNotificationsKey(userID.String(), limit, offset))
	if err == nil {
		return *notificationsCache, nil
	}
	if err != redis.Nil {
		s.logger.Sugar().Errorf("failed to get user(%s)'s notifications from redis: %s", userID.String(), err.Error())
		return nil, ErrInternal
	}

	notifications, err := s.repo.Postgres.Notification.GetUserNotifications(ctx, userID, limit, offset)
	if err != nil && err != pgx.ErrNoRows {
		s.logger.Sugar().Errorf("failed to get user(%s)'s notifications from postgres: %s", userID.String(), err.Error())
		return nil, ErrInternal
	}

	if err := redisrepo.SetJSON(s.rdb, ctx, redisrepo.UserNotificationsKey(userID.String(), limit, offset), notifications, time.Minute * 2); err != nil {
		s.logger.Sugar().Errorf("failed to set user(%s)'s notification in redis cache: %s", userID.String(), err.Error())
	}

	return notifications, nil
}
