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
	OLD_NOTIFICATIONS_DAYS = 14
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

	query := "INSERT INTO notifications(type, receiver_id, content, resource_id) VALUES "
	values := []interface{}{}
	counter := 1

	for _, n := range notifications {
		query += fmt.Sprintf("($%d, $%d, $%d, $%d),", counter, counter+1, counter+2, counter+3)
		values = append(values, n.Type, n.ReceiverID, n.Content, n.ResourceID)
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

func (r *notificationRepo) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Notification, error) {
	if limit > GET_NOTIFICATIONS_MAX_LIMIT {
		limit = GET_NOTIFICATIONS_MAX_LIMIT
	}
	
	rows, err := r.db.Query(
		ctx,
		`
		SELECT n.id, n.type, n.content, n.resource_id, n.created_at
		FROM notifications n
		WHERE n.receiver_id = $1
		ORDER BY n.created_at DESC
		LIMIT $2
		OFFSET $3
		`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*model.Notification
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.Type, &n.Content, &n.ResourceID, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.ReceiverID = userID

		notifications = append(notifications, &n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *notificationRepo) DeleteOldNotifications(ctx context.Context) error {
	_, err := r.db.Exec(ctx, "DELETE FROM notifications WHERE created_at < NOW() - MAKE_INTERVAL(days => $1)", OLD_NOTIFICATIONS_DAYS)
	return err
}

func (r *notificationRepo) CreateGlobalNotification(ctx context.Context, gn model.GlobalNotification) error {
	_, err := r.db.Exec(ctx, "INSERT INTO global_notifications(id, poster_id, title, content, resource_link) VALUES($1, $2, $3, $4, $5)", gn.ID, gn.PosterID, gn.Title, gn.Content, gn.ResourceLink)
	return err
}

func (r *notificationRepo) GetGlobalNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.GlobalNotification, error) {
	if limit > GET_NOTIFICATIONS_MAX_LIMIT {
		limit = GET_NOTIFICATIONS_MAX_LIMIT
	}

	rows, err := r.db.Query(
		ctx,
		`
		SELECT g.id, g.poster_id, g.title, g.content, g.resource_link, g.created_at
		FROM global_notifications g
		LEFT JOIN checked_global_notifications c
			ON c.user_id = $1 AND c.notification_id = g.id
		WHERE c.notification_id IS NULL
		ORDER BY g.created_at DESC
		LIMIT $2
		OFFSET $3
		`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*model.GlobalNotification
	for rows.Next() {
		var n model.GlobalNotification
		if err := rows.Scan(&n.ID, &n.PosterID, &n.Title, &n.Content, &n.ResourceLink, &n.CreatedAt); err != nil {
			return nil, err
		}

		notifications = append(notifications, &n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *notificationRepo) MarkGlobalNotificationAsRead(ctx context.Context, userID uuid.UUID, notificationID int64) error {
	_, err := r.db.Exec(ctx, "INSERT INTO checked_global_notifications(user_id, notification_id) VALUES($1, $2)", userID, notificationID)
	return err
}
