package handler

import (
	"encoding/json"
	"net/http"

	"github.com/BloggingApp/notification-service/internal/service"
)

type Resp map[string]interface{}

type Handler struct {
	services *service.Service
}

func New(services *service.Service) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// GET
	mux.HandleFunc("/api/v1/notifications", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		user, err := h.authMiddleware(r)
		if err != nil {
			h.Respond(w, Resp{"error": err.Error()}, http.StatusUnauthorized)
			return
		}

		h.notificationsGet(user, w, r)
	})

	return mux
}

func (h *Handler) Respond(w http.ResponseWriter, resp any, statusCode int) {
	respJSON, _ := json.Marshal(resp)
	w.WriteHeader(statusCode)
	w.Write(respJSON)
}
