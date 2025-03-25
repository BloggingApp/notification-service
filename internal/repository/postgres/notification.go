package postgres

import (
	"context"
	"fmt"

	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	GET_NOTIFICATIONS_MAX_LIMIT = 10
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
	_, err := r.db.Exec(ctx, query, values...)
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

func (r *notificationRepo) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*model.Notification, error) {
	if limit > GET_NOTIFICATIONS_MAX_LIMIT {
		limit = GET_NOTIFICATIONS_MAX_LIMIT
	}
	
	rows, err := r.db.Query(
		ctx,
		`
		SELECT n.id, n.type, n.message, n.created_at
		FROM notifications n
		WHERE n.receiver_id = $1
		LIMIT $2
		OFFSET $3
		ORDER BY n.created_at DESC
		`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*model.Notification
	for rows.Next() {
		var notification model.Notification
		if err := rows.Scan(&notification.ID, &notification.Type, &notification.Message, &notification.CreatedAt); err != nil {
			return nil, err
		}
		notification.ReceiverID = userID

		notifications = append(notifications, &notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}
