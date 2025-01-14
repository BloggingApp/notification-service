package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BloggingApp/notification-service/internal/mailer"
	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
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

	mailer := mailer.New(logger, rabbitmq)

	mailer.StartProcessing()

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
