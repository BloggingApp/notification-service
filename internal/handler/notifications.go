package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BloggingApp/notification-service/internal/dto"
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

func (h *Handler) notificationsCreateManually(admin *model.User, w http.ResponseWriter, r *http.Request) {
	if admin == nil {
		return
	}

	var input dto.CreateNotificationManually
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.Respond(w, Resp{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	if err := h.services.Notification.CreateGlobalNotification(r.Context(), model.GlobalNotification{
		PosterID: admin.ID,
		Title: input.Title,
		Content: input.Content,
		ResourceLink: input.ResourceLink,
	}); err != nil {
		h.Respond(w, Resp{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	h.Respond(w, Resp{}, http.StatusCreated)
}

func (h *Handler) notificationsGetGlobal(user *model.User, w http.ResponseWriter, r *http.Request) {
	if user == nil {
		return
	}

	limitString := r.URL.Query().Get("limit")
	offsetString := r.URL.Query().Get("offset")
	limit, err0 := strconv.Atoi(limitString)
	offset, err1 := strconv.Atoi(offsetString)
	if err0 != nil || err1 != nil {
		h.Respond(w, Resp{"error": errInvalidLimitOffset.Error()}, http.StatusBadRequest)
		return
	}

	notifications, err := h.services.Notification.GetGlobalNotifications(r.Context(), user.ID, limit, offset)
	if err != nil {
		h.Respond(w, Resp{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	h.Respond(w, notifications, http.StatusOK)
}

func (h *Handler) notificationsMarkGlobalNotificationAsRead(user *model.User, w http.ResponseWriter, r *http.Request) {
	if user == nil {
		return
	}

	notificationIDString := r.PathValue("nId")
	if notificationIDString == "" {
		return
	}
	notificationID, err := strconv.Atoi(notificationIDString)
	if err != nil {
		h.Respond(w, Resp{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	if err := h.services.Notification.MarkGlobalNotificationAsRead(r.Context(), user.ID, int64(notificationID)); err != nil {
		h.Respond(w, Resp{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	h.Respond(w, Resp{}, http.StatusOK)
}
