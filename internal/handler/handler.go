package handler

import (
	"net/http"

	"github.com/BloggingApp/notification-service/internal/service"
)

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

	return mux
}
