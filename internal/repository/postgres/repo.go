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
}

type PGRepo struct {
	User
}

func New(db *pgxpool.Conn) *PGRepo {
	return &PGRepo{
		User: newUserRepo(db),
	}
}
