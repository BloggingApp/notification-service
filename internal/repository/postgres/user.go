package postgres

import (
	"context"
	"strconv"

	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	db *pgxpool.Pool
}

func newUserRepo(db *pgxpool.Pool) User {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) Create(ctx context.Context, user model.User) error {
	_, err := r.db.Exec(ctx, "INSERT INTO users(id, username, display_name) VALUES($1, $2, $3)", user.ID, user.Username, user.DisplayName)
	return err
}

func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.QueryRow(ctx, "SELECT u.id, u.username, u.display_name FROM users u WHERE u.id = $1", id).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) UpdateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	query := "UPDATE users SET "
	args := []interface{}{}
	i := 1

	for column, value := range updates {
		query += (column + " = $" + strconv.Itoa(i) + ", ")
		args = append(args, value)
		i++
	}

	query = query[:len(query)-2] + " WHERE id = $" + strconv.Itoa(i) + " RETURNING id"
	args = append(args, id)

	var returnedID uuid.UUID
	err := r.db.QueryRow(ctx, query, args...).Scan(&returnedID)
	if err == pgx.ErrNoRows {
		return nil
	}
	return err
}
