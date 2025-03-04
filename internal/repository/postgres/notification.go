package postgres

import (
	"context"
	"fmt"

	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notificationRepo struct {
	db *pgxpool.Pool
}

func newNotificationRepo(db *pgxpool.Pool) Notification {
	return &notificationRepo{
		db: db,
	}
}

func (r *notificationRepo) GetInterestedFollowers(ctx context.Context, authorID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		SELECT follower_id FROM followers
		WHERE user_id = $1 AND new_post_notifications_enabled = true
	`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followerIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		followerIDs = append(followerIDs, id)
	}

	return followerIDs, nil
}

func (r *notificationRepo) CreateBatch(ctx context.Context, notifications []model.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	query := "INSERT INTO notifications(type, receiver_id, message, created_at) VALUES "
	values := []interface{}{}
	counter := 1

	for _, n := range notifications {
		query += fmt.Sprintf("($%d, $%d, $%d, NOW()),", counter, counter+1, counter+2)
		values = append(values, n.Type, n.ReceiverID, n.Message)
		counter += 3
	}

	query = query[:len(query)-1]
	_, err := r.db.Exec(ctx, query)
	return err
}

func (r *notificationRepo) CreateBatched(ctx context.Context, notifications []model.Notification, batchSize int) error {
	for i := 0; i < len(notifications); i += batchSize {
		end := i + batchSize
		if end > len(notifications) {
			end = len(notifications)
		}

		if err := r.CreateBatch(ctx, notifications[i:end]); err != nil {
			return err
		}
	}

	return nil
}
