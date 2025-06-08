package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/BloggingApp/notification-service/internal/dto"
	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/BloggingApp/notification-service/internal/repository"
	"github.com/BloggingApp/notification-service/internal/repository/redisrepo"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type notificationService struct {
	logger *zap.Logger
	repo *repository.Repository
	rdb *redis.Client
	rabbitmq *rabbitmq.MQConn
	scheduler gocron.Scheduler
	conns *sync.Map
	deliveryChan chan model.NotificationDelivery
}

func newNotificationService(logger *zap.Logger, repo *repository.Repository, rdb *redis.Client, rabbitmq *rabbitmq.MQConn) Notification {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	s := &notificationService{
		logger: logger,
		repo: repo,
		rdb: rdb,
		rabbitmq: rabbitmq,
		scheduler: scheduler,
		conns: &sync.Map{},
		deliveryChan: make(chan model.NotificationDelivery, 1000),
	}

	for range 5 {
		go s.deliveryWorker()
	}

	return s
}

func (s *notificationService) deliveryWorker() {
	for msg := range s.deliveryChan {
		val, ok := s.conns.Load(msg.ReceiverID)
		if !ok {
			continue
		}

		conn, ok := val.(*websocket.Conn)
		if !ok {
			continue
		}

		payload := map[string]string{
			"type": msg.Type,
			"content": msg.Content,
			"resource_id": msg.ResourceID,
		}
		if err := conn.WriteJSON(payload); err != nil {
			s.logger.Sugar().Errorf("failed to write json msg to receiver(%s)'s conn: %s", msg.ReceiverID.String(), err.Error())
		}
	}
}

func (s *notificationService) RegisterConnection(userID uuid.UUID, conn *websocket.Conn) {
	s.conns.Store(userID, conn)

	go func(userID uuid.UUID, c *websocket.Conn) {
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				s.UnregisterConnection(userID)
				break
			}
		}
	}(userID, conn)
}

func (s *notificationService) UnregisterConnection(userID uuid.UUID) {
	if val, ok := s.conns.Load(userID); ok {
		if conn, ok := val.(*websocket.Conn); ok {
			conn.Close()
		}
		s.conns.Delete(userID)
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
		content := fmt.Sprintf("%s has created new post: %s", author.Username, postCreatedDto.PostTitle)
		resourceID := strconv.Itoa(int(postCreatedDto.PostID))

		var notifications []model.Notification
		for _, receiver := range receivers {
			notifications = append(notifications, model.Notification{
				Type: notificationType,
				ReceiverID: receiver,
				Content: content,
				ResourceID: resourceID,
			})
		}

		if err := s.repo.Postgres.Notification.CreateBatched(ctx, notifications, 1000); err != nil {
			s.logger.Sugar().Errorf("failed to create batched notifications for post(%d): %s", postCreatedDto.PostID, err.Error())
			msg.Ack(false)
			continue
		}

		msg.Ack(false)

		for _, receiver := range receivers {
			s.deliveryChan <- model.NotificationDelivery{
				ReceiverID: receiver,
				Type: notificationType,
				Content: content,
				ResourceID: resourceID,
			}
		}
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

func (s *notificationService) newDeleteOldNotificationsJob() {
	s.scheduler.NewJob(gocron.DurationJob(time.Hour * 12), gocron.NewTask(func(ctx context.Context) {
		if err := s.repo.Postgres.Notification.DeleteOldNotifications(ctx); err != nil {
			s.logger.Sugar().Errorf("failed to delete old notifications: %s", err.Error())
		}
	}))
}

func (s *notificationService) StartJobs() {
	s.newDeleteOldNotificationsJob()

	s.scheduler.Start()
}

func (s *notificationService) CreateGlobalNotification(ctx context.Context, gn model.GlobalNotification) error {
	if len(gn.Title) > 255 || gn.Title == "" || len(gn.ResourceLink) > 255 {
		return ErrInvalidInputForGlobalNotification
	}

	return s.repo.Postgres.Notification.CreateGlobalNotification(ctx, gn)
}

func (s *notificationService) GetGlobalNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.GlobalNotification, error) {
	return s.repo.Postgres.Notification.GetGlobalNotifications(ctx, userID, limit, offset)
}

func (s *notificationService) MarkGlobalNotificationAsRead(ctx context.Context, userID uuid.UUID, notificationID int64) error {
	return s.repo.Postgres.Notification.MarkGlobalNotificationAsRead(ctx, userID, notificationID)
}
