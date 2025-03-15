package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BloggingApp/notification-service/internal/config"
	"github.com/BloggingApp/notification-service/internal/mailer"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/BloggingApp/notification-service/internal/repository"
	"github.com/BloggingApp/notification-service/internal/repository/postgres"
	"github.com/BloggingApp/notification-service/internal/service"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	if err := initEnv(); err != nil {
		log.Fatalf("failed to load environment variables: %s", err.Error())
	}

	logger, err := newLogger()
	if err != nil {
		log.Fatalf("failed to create zap logger: %s", err.Error())
	}

	rabbitmq, err := rabbitmq.New(os.Getenv("RABBITMQ_CONN_STRING"))
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %s", err.Error())
	}

	db, err := postgres.Connect(ctx, config.DBConfig{
		Username: os.Getenv("POSTGRES_USERNAME"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Host: os.Getenv("POSTGRES_HOST"),
		Port: os.Getenv("POSTGRES_PORT"),
		DBName: os.Getenv("POSTGRES_DB"),
		SSLMode: os.Getenv("POSTGRES_SSLMODE"),
	})
	if err != nil {
		log.Panicf("db connection error: %s", err.Error())
	}
	if err := db.Ping(ctx); err != nil {
		log.Panicf("couldn't ping postgres db: %s", err.Error())
	}
	log.Println("Successfully connected to PostgreSQL")

	repos := repository.New(db)
	services := service.New(logger, repos, rabbitmq)

	mailer := mailer.New(logger, rabbitmq)
	mailer.StartProcessing()

	services.User.StartCreating(ctx)
	services.User.StartUpdating(ctx)
	services.Notification.StartProcessingNewPostNotifications(ctx)

	log.Println("Notification service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Notification service shutting Down")
}

func initEnv() error {
	return godotenv.Load(".env")
}

func newLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"./app.log",
	}
	return cfg.Build()
}
