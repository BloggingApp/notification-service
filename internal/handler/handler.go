package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/BloggingApp/notification-service/internal/service"
	"github.com/golang-jwt/jwt/v5"
	jwtmanager "github.com/morf1lo/jwt-pair-manager"
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

	mux.HandleFunc("/api/v1/notifications/global", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			admin, err := h.adminMiddleware(r)
			if err != nil {
				h.Respond(w, Resp{"error": err.Error()}, http.StatusForbidden)
				return
			}

			h.notificationsCreateManually(admin, w, r)
		} else if r.Method == http.MethodGet {
			user, err := h.authMiddleware(r)
			if err != nil {
				h.Respond(w, Resp{"error": err.Error()}, http.StatusForbidden)
				return
			}

			h.notificationsGetGlobal(user, w, r)
		}
	})

	mux.HandleFunc("/api/v1/notifications/global/{nId}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			return
		}

		user, err := h.authMiddleware(r)
		if err != nil {
			h.Respond(w, Resp{"error": err.Error()}, http.StatusForbidden)
			return
		}

		h.notificationsMarkGlobalNotificationAsRead(user, w, r)
	})

	return mux
}

func (h *Handler) Respond(w http.ResponseWriter, resp any, statusCode int) {
	respJSON, _ := json.Marshal(resp)
	w.WriteHeader(statusCode)
	w.Write(respJSON)
}

func (h *Handler) GetJWTClaimsFromRequest(r *http.Request) (jwt.MapClaims, error) {
	bearerHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(bearerHeader, "Bearer ") {
		return nil, errInvalidJWT
	}

	token := strings.Split(bearerHeader, " ")[1]
	if token == "" {
		return nil, errNoToken
	}

	claims, err := jwtmanager.DecodeJWT(token, []byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	return claims, nil
}
