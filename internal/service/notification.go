package service

import (
	"context"
	"encoding/json"

	"github.com/BloggingApp/notification-service/internal/dto"
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

		
	}
}
