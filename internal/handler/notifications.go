package handler

import (
	"net/http"
	"strconv"

	"github.com/BloggingApp/notification-service/internal/model"
)

func (h *Handler) notificationsGet(user *model.User, w http.ResponseWriter, r *http.Request) {
	if user == nil {
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		h.Respond(w, Resp{"error": "'limit' parameter isn't set or isn't a number."}, http.StatusBadRequest)
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		h.Respond(w, Resp{"error": "'offset' parameter isn't set or isn't a number."}, http.StatusBadRequest)
		return
	}

	notifications, err := h.services.Notification.GetUserNotifications(r.Context(), user.ID, limit, offset)
	if err != nil {
		h.Respond(w, Resp{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	h.Respond(w, notifications, http.StatusOK)
}
