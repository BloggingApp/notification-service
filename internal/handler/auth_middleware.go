package handler

import (
	"net/http"

	"github.com/BloggingApp/notification-service/internal/model"
	"github.com/google/uuid"
)

func (h *Handler) authMiddleware(r *http.Request) (*model.User, error) {
	claims, err := h.GetJWTClaimsFromRequest(r)
	if err != nil {
		return nil, err
	}

	userIDString, ok := claims["id"].(string)
	if !ok {
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
