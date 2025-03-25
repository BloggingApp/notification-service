package redisrepo

import "fmt"

const (
	USER_NOTIFICATIONS = "user:%s-notifications:%d:%d" // <userID>:<limit>:<offset>
)

func UserNotificationsKey(userID string, limit int, offset int) string {
	return fmt.Sprintf(USER_NOTIFICATIONS, userID, limit, offset)
}
