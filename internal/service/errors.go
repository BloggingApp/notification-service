package service

import "errors"

var (
	ErrInternal = errors.New("internal server error")
	ErrInvalidInputForGlobalNotification = errors.New("title and resource_link must not be over 255. and title is required")
)
