package handler

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/google/uuid"
	jwtmanager "github.com/morf1lo/jwt-pair-manager"
)

var (
	errNoToken = errors.New("there is no token")
	errInvalidJWT = errors.New("invalid jwt")
	errInvalidUserID = errors.New("invalid user ID")
)

func (h *Handler) authMiddleware(r *http.Request) (*model.User, error) {
	bearerHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(bearerHeader, "Bearer ") {
		return nil, errNoToken
	}

	token := strings.Split(bearerHeader, " ")[1]
	if token == "" {
		return nil, errNoToken
	}

	claims, err := jwtmanager.DecodeJWT(token, []byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	userIDString, exists := claims["id"].(string)
	if !exists {
		return nil, errInvalidJWT
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return nil, errInvalidUserID
	}

	user, err := h.services.User.FindByID(r.Context(), userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
