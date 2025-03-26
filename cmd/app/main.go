package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BloggingApp/notification-service/internal/config"
	"github.com/BloggingApp/notification-service/internal/handler"
	"github.com/BloggingApp/notification-service/internal/mailer"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/BloggingApp/notification-service/internal/repository"
	"github.com/BloggingApp/notification-service/internal/repository/postgres"
	"github.com/BloggingApp/notification-service/internal/service"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	if err := loadEnv(); err != nil {
		log.Fatalf("failed to load environment variables: %s", err.Error())
	}

	if err := initConfig(); err != nil {
		log.Fatalf("failed to initialize config: %s", err.Error())
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
		Username: os.Getenv("POSTGRES_USER"),
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

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic("failed to ping redis: " + err.Error())
	}
	log.Printf("Successfully connected to Redis: %s\n", pong)

	repo := repository.New(db)
	services := service.New(logger, repo, rdb, rabbitmq)
	handlers := handler.New(services)

	mailer := mailer.New(logger, rabbitmq)
	mailer.StartProcessing()

	go services.User.StartCreating(ctx)
	go services.User.StartUpdating(ctx)
	go services.User.StartCreatingFollowers(ctx)
	go services.Notification.StartProcessingNewPostNotifications(ctx)

	go services.Notification.StartJobs()

	go http.ListenAndServe(viper.GetString("app.port"), handlers.SetupRoutes())

	log.Println("Notification service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Notification service shutting Down")
}

func loadEnv() error {
	return godotenv.Load(".env")
}

func initConfig() error {
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("app")
	return viper.ReadInConfig()
}

func newLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"./app.log",
	}
	return cfg.Build()
}
