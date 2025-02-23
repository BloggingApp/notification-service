package service

import (
	"context"

	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/BloggingApp/notification-service/internal/repository"
	"github.com/google/uuid"
)

type User interface {
	create(ctx context.Context, user model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	updateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	StartUpdating(ctx context.Context)
}

type Service struct {
	User
}

func New(repo *repository.Repository, rabbitmq *rabbitmq.MQConn) *Service {
	return &Service{

	}
}
