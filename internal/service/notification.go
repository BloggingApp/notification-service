package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/BloggingApp/notification-service/internal/dto"
	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/BloggingApp/notification-service/internal/repository"
	"go.uber.org/zap"
)

type notificationService struct {
	logger *zap.Logger
	repo *repository.Repository
	rabbitmq *rabbitmq.MQConn
}

func newNotificationService(logger *zap.Logger, repo *repository.Repository, rabbitmq *rabbitmq.MQConn) Notification {
	return &notificationService{
		logger: logger,
		repo: repo,
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
