package service

import (
	"context"
	"encoding/json"

	"github.com/BloggingApp/notification-service/internal/dto"
	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/BloggingApp/notification-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type userService struct {
	logger *zap.Logger
	repo *repository.Repository
	rabbitmq *rabbitmq.MQConn
}

func newUserService(logger *zap.Logger, repo *repository.Repository, rabbitmq *rabbitmq.MQConn) User {
	return &userService{
		logger: logger,
		repo: repo,
		rabbitmq: rabbitmq,
	}
}

func (s *userService) create(ctx context.Context, user model.User) error {
	return s.repo.Postgres.User.Create(ctx, user)
}

// TODO: cache result
func (s *userService) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.repo.Postgres.User.FindByID(ctx, id)	
}

func (s *userService) updateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	allowedFields := []string{"username", "display_name", "avatar_url"}
	allowedFieldsSet := make(map[string]struct{}, len(allowedFields))
	for _, field := range allowedFields {
		allowedFieldsSet[field] = struct{}{}
	}

	for field := range updates {
		if _, ok := allowedFieldsSet[field]; !ok {
			delete(updates, field)
		}
	}

	if len(updates) == 0 {
		return nil
	}

	return s.repo.Postgres.User.UpdateByID(ctx, id, updates)
}

func (s *userService) StartCreating(ctx context.Context) {
	msgs, err := s.rabbitmq.ConsumeExchange(rabbitmq.USERS_CREATED_EXCHANGE)
	if err != nil {
		panic(err)
	}

	for msg := range msgs {
		var userCreatedDto dto.MQUserCreated
		if err := json.Unmarshal(msg.Body, &userCreatedDto); err != nil {
			msg.Ack(false)
			continue
		}

		if err := s.create(ctx, model.User{
			ID: userCreatedDto.ID,
			Username: userCreatedDto.Username,
			DisplayName: userCreatedDto.DisplayName,
			AvatarURL: userCreatedDto.AvatarURL,
		}); err != nil {
			s.logger.Sugar().Errorf("failed to create user(%s): %s", userCreatedDto.ID.String(), err.Error())
			msg.Ack(false)
			continue
		}

		msg.Ack(false)
	}
}

func (s *userService) StartUpdating(ctx context.Context) {
	msgs, err := s.rabbitmq.ConsumeExchange(rabbitmq.USERS_UPDATE_EXCHANGE)
	if err != nil {
		panic(err)
	}
	
	for msg := range msgs {
		var updates map[string]interface{}
		if err := json.Unmarshal(msg.Body, &updates); err != nil {
			msg.Ack(false)
			continue
		}

		userIDString, ok := updates["user_id"].(string)
		if !ok {
			msg.Ack(false)
			continue
		}
		userID, err := uuid.Parse(userIDString)
		if err != nil {
			msg.Ack(false)
			continue
		}

		delete(updates, "user_id")

		if err := s.updateByID(ctx, userID, updates); err != nil {
			s.logger.Sugar().Errorf("failed to update user(%s): %s", userID.String(), err.Error())
			msg.Ack(false)
			continue
		}

		msg.Ack(false)
	}
}
