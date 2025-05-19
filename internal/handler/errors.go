package handler

import "errors"

var (
	errNoToken              = errors.New("there is no token")
	errInvalidJWT           = errors.New("invalid jwt")
	errInvalidUserID        = errors.New("invalid user ID")
	errNotAdmin             = errors.New("you are not an admin")
	errInvalidLimitOffset = errors.New("limit and offset must be integer")
)
