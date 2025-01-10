package mailer

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/BloggingApp/notification-service/internal/rabbitmq"
	"go.uber.org/zap"
)

type Mailer struct {
	logger *zap.Logger
	rabbitmq *rabbitmq.MQConn

	from string
	pass string
	host string
	port string
}

func New(logger *zap.Logger, rabbitmq *rabbitmq.MQConn) *Mailer {
	return &Mailer{
		rabbitmq: rabbitmq,
		from: os.Getenv("FROM"),
		pass: os.Getenv("PASS"),
		host: os.Getenv("HOST"),
		port: os.Getenv("PORT"),
	}
}

func (m *Mailer) StartProcessing() {
	go m.ProcessRegistrationCodes()
	go m.ProcessSignInCodes()
}

func (m *Mailer) ProcessRegistrationCodes() {
	queue := rabbitmq.REGISTRATION_CODE_MAIL_QUEUE
	msgs, err := m.rabbitmq.Consume(queue)
	if err != nil {
		m.logger.Sugar().Fatalf("Failed to start consuming(%s): %s", queue, err.Error())
	}

	for msg := range msgs {
		var message rabbitmq.NotificateUserCodeDto
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			m.logger.Sugar().Errorf("Failed to unmarshal json in queue(%s): %s", queue, err.Error())
			msg.Nack(false, false)
			continue
		}

		if err := m.SendRegistrationCodeMail(message); err != nil {
			m.logger.Sugar().Errorf("Failed to send mail to(%s): %s", message.Email, err.Error())
			msg.Nack(false, true)
			continue
		}

		msg.Ack(false)

		m.logger.Sugar().Infof("Successfully sent registration code from queue(%s) to(%s)", queue, message.Email)
		time.Sleep(time.Millisecond * 10)
	}
}

func (m *Mailer) SendRegistrationCodeMail(input rabbitmq.NotificateUserCodeDto) error {
	subject := "Verify your email"
	body := fmt.Sprintf("Your code:\n<b>%d</b>", input.Code)

	msg := []byte("Subject: " + subject + "\r\n" +
	"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
	"\r\n" + body)

	auth := smtp.PlainAuth("", m.from, m.pass, m.host)

	if err := smtp.SendMail(m.host + ":" + m.port, auth, m.from, []string{input.Email}, msg); err != nil {
		return err
	}

	return nil
}

func (m *Mailer) ProcessSignInCodes() {
	queue := rabbitmq.SIGNIN_CODE_MAIL_QUEUE
	msgs, err := m.rabbitmq.Consume(queue)
	if err != nil {
		m.logger.Sugar().Fatalf("Failed to start consuming(%s): %s", queue, err.Error())
	}

	for msg := range msgs {
		var message rabbitmq.NotificateUserCodeDto
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			m.logger.Sugar().Errorf("Failed to unmarshal json in queue(%s): %s", queue, err.Error())
			msg.Nack(false, false)
			continue
		}

		if err := m.SendSignInCodeMail(message); err != nil {
			m.logger.Sugar().Errorf("Failed to send mail to(%s): %s", message.Email, err.Error())
			msg.Nack(false, true)
			continue
		}

		msg.Ack(false)

		m.logger.Sugar().Infof("Successfully sent auth code from queue(%s) to(%s)", queue, message.Email)
		time.Sleep(time.Millisecond * 10)
	}
}

func (m *Mailer) SendSignInCodeMail(input rabbitmq.NotificateUserCodeDto) error {
	subject := "Two-factor authentication"
	body := fmt.Sprintf("Your code:\n<b>%d</b>", input.Code)

	msg := []byte("Subject: " + subject + "\r\n" +
	"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
	"\r\n" + body)

	auth := smtp.PlainAuth("", m.from, m.pass, m.host)

	if err := smtp.SendMail(m.host + ":" + m.port, auth, m.from, []string{input.Email}, msg); err != nil {
		return err
	}

	return nil
}
