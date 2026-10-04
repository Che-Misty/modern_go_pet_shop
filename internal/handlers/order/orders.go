package order

import (
	"context"
	"go-pet-shop/internal/models"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Orders interface {
	CreateOrder(ctx context.Context, order models.Order) (int, error)
	AddOrderItem(ctx context.Context, orderItem models.OrderItem) error
	GetOrderByID(ctx context.Context, id int) (*models.Order, error)
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error)
}

type Handler struct {
	log     *slog.Logger
	storage Orders
}

func NewHandler(log *slog.Logger, storage Orders) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.order.CreateOrder"
	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Creating new order", slog.String("url", r.URL.String()))

	var order models.Order
	if err := render.DecodeJSON(r.Body, &order); err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid JSON payload",
		})
		return
	}

	orderID, err := h.storage.CreateOrder(r.Context(), order)
	if err != nil {
		log.Error("failed to create order", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to create order",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"status":   "Order created successfully",
		"order_id": orderID,
		"order":    order,
	})
}

func (h *Handler) AddOrderItem(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.order.AddOrderItem"
	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		log.Error("missing order ID parameter")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Missing order ID parameter",
		})
		return
	}

	id, err := strconv.Atoi(orderID)
	if err != nil {
		log.Error("invalid order ID parameter", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid order ID parameter",
		})
		return
	}

	log.Info("Adding order item", slog.String("url", r.URL.String()))

	var orderItem models.OrderItem
	if err := render.DecodeJSON(r.Body, &orderItem); err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid JSON payload",
		})
		return
	}

	orderItem.OrderID = id

	err = h.storage.AddOrderItem(r.Context(), orderItem)
	if err != nil {
		log.Error("failed to add order item", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to add order item",
		})
		return
	}
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message": "Order item added successfully",
	})
}

func (h *Handler) GetOrdersByUserEmail(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.order.GetOrdersByUserEmail"
	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	email := r.URL.Query().Get("email")
	if email == "" {
		log.Error("missing user email parameter")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Missing user email parameter",
		})
		return
	}

	orders, err := h.storage.GetOrdersByUserEmail(r.Context(), email)
	if err != nil {
		log.Error("failed to get orders", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to get orders",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, orders)
}

type OrderDetailResponse struct {
	Order      models.Order       `json:"order"`
	OrderItems []models.OrderItem `json:"order_items"`
}

func (h *Handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.order.GetOrderByID"
	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		log.Error("missing order ID parameter")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Missing order ID parameter",
		})
		return
	}

	id, err := strconv.Atoi(orderID)
	if err != nil {
		log.Error("invalid order ID", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid order ID",
		})
		return
	}

	order, err := h.storage.GetOrderByID(r.Context(), id)
	if err != nil {
		log.Error("failed to get order", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to get order",
		})
		return
	}

	orderItems, err := h.storage.GetOrderItemsByOrderID(r.Context(), id)
	if err != nil {
		log.Error("failed to get order items", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to get order items",
		})
		return
	}

	response := OrderDetailResponse{
		Order:      *order,
		OrderItems: orderItems,
	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response)
}
