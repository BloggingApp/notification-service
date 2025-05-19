package dto

type CreateNotificationManually struct {
	Title        string `json:"title" binding:"required,max=255"`
	Content      string `json:"content" binding:"required"`
	ResourceLink string `json:"resource_link" binding:"max=255"`
}
