package user

import (
	"context"
	"go-pet-shop/internal/models"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type User interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type Handler struct {
	log     *slog.Logger
	storage User
}

func New(log *slog.Logger, storage User) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.user.CreateUser"
	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Creating new user", slog.String("url", r.URL.String()))

	var user models.User
	if err := render.DecodeJSON(r.Body, &user); err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid JSON payload",
		})
		return
	}

	if user.Name == "" {
		log.Error("username is empty")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Username is required",
		})
		return
	}

	if user.Email == "" {
		log.Error("user email is empty")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Email is required",
		})
		return
	}

	userID, err := h.storage.CreateUser(r.Context(), user)

	if err != nil {
		log.Error("failed to create user", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to create user",
		})
		return
	}

	log.Info("User created successfully",
		slog.Int("id", userID),
		slog.String("name", user.Name),
		slog.String("url", r.URL.String()),
	)

	user.ID = userID
	render.JSON(w, r, map[string]interface{}{
		"status": "User created successfully",
		"id":     userID,
		"user":   user,
	})

}

func (h *Handler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.user.GetUserByEmail"
	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	email := chi.URLParam(r, "email")
	if email == "" {
		log.Error("email query parameter is required")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Email is required",
		})
		return
	}

	user, err := h.storage.GetUserByEmail(r.Context(), email)
	if err != nil {
		log.Error("failed to retrieve user", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to retrieve user",
		})
		return
	}

	if user == nil {
		log.Info("User not found", slog.String("email", email))
		w.WriteHeader(http.StatusNotFound)
		render.JSON(w, r, map[string]string{
			"error":   "Not found",
			"message": "User not found",
		})
		return
	}

	log.Info("Retrieved user successfully",
		slog.String("url", r.URL.String()),
		slog.Int("id", user.ID),
	)

	render.JSON(w, r, user)
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.user.GetAllUsers"
	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Retrieving all users", slog.String("url", r.URL.String()))

	users, err := h.storage.GetAllUsers(r.Context())
	if err != nil {
		log.Error("failed to retrieve users", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to retrieve users",
		})
		return
	}

	log.Info("Retrieved all users successfully",
		slog.String("url", r.URL.String()),
		slog.Int("count", len(users)),
	)

	render.JSON(w, r, users)
}
