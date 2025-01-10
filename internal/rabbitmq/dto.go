package rabbitmq

type NotificateUserCodeDto struct {
	Email string `json:"email"`
	Code  int    `json:"code"`
}
