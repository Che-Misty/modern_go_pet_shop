package order

import (
	"bytes"
	"context"
	"encoding/json"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateOrder_Success(t *testing.T) {
	newOrder := models.Order{
		ID:         1,
		UserEmail:  "john.doe@example.com",
		TotalPrice: 100.0,
	}

	jsonData, err := json.Marshal(newOrder)
	if err != nil {
		t.Fatal(err)
	}

	CreateOrderMock := NewOrdersMock(t)
	CreateOrderMock.EXPECT().CreateOrder(mock.Anything, newOrder).Return(1, nil)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), CreateOrderMock)
	handler.CreateOrder(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateOrder_BadRequest(t *testing.T) {
	jsonData := []byte(`{"id": 1, "user_email": "john.doe@example.com", "total_price": 100.0`)

	CreateOrderMock := NewOrdersMock(t)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), CreateOrderMock)
	handler.CreateOrder(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_Fail(t *testing.T) {
	newOrder := models.Order{
		ID:         1,
		UserEmail:  "john.doe@example.com",
		TotalPrice: 100.0,
	}

	jsonData, err := json.Marshal(newOrder)
	if err != nil {
		t.Fatal(err)
	}

	CreateOrderMock := NewOrdersMock(t)
	CreateOrderMock.EXPECT().CreateOrder(mock.Anything, newOrder).Return(0, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), CreateOrderMock)
	handler.CreateOrder(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAddOrderItem_Success(t *testing.T) {
	newOrderItem := models.OrderItem{
		ID:        1,
		OrderID:   1,
		ProductID: 1,
		Quantity:  2,
	}

	jsonData, err := json.Marshal(newOrderItem)
	if err != nil {
		t.Fatal(err)
	}

	AddOrderItemMock := NewOrdersMock(t)
	AddOrderItemMock.EXPECT().AddOrderItem(mock.Anything, newOrderItem).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/orders/1/items", bytes.NewReader(jsonData))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), AddOrderItemMock)
	handler.AddOrderItem(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestAddOrderItem_BadRequest(t *testing.T) {
	newOrderItem := []byte(`{"id": 1, "order_id": 1, "product_id": 1, "quantity": 2`)

	AddOrderItemMock := NewOrdersMock(t)

	req := httptest.NewRequest(http.MethodPost, "/orders/1/items", bytes.NewReader(newOrderItem))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), AddOrderItemMock)
	handler.AddOrderItem(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddOrderItem_Fail(t *testing.T) {
	newOrderItem := models.OrderItem{
		ID:        2,
		OrderID:   3,
		ProductID: 3,
		Quantity:  5,
	}

	AddOrderItemMock := NewOrdersMock(t)
	AddOrderItemMock.EXPECT().AddOrderItem(mock.Anything, newOrderItem).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/orders/3/items", bytes.NewReader([]byte(`{"id": 2, "order_id": 3, "product_id": 3, "quantity": 5}`)))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "3")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), AddOrderItemMock)
	handler.AddOrderItem(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetOrdersByUserEmail_Success(t *testing.T) {
	GetOrdersByUserEmailMock := NewOrdersMock(t)
	GetOrdersByUserEmailMock.EXPECT().GetOrdersByUserEmail(mock.Anything, "test@example.com").Return([]models.Order{{ID: 1, UserEmail: "test@example.com"}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/orders/user/?email=test@example.com", nil)
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), GetOrdersByUserEmailMock)
	handler.GetOrdersByUserEmail(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetOrdersByUserEmail_BadRequest(t *testing.T) {
	GetOrdersByUserEmailMock := NewOrdersMock(t)

	req := httptest.NewRequest(http.MethodGet, "/orders/user/", nil)
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), GetOrdersByUserEmailMock)
	handler.GetOrdersByUserEmail(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetOrdersByUserEmail_Fail(t *testing.T) {
	GetOrdersByUserEmailMock := NewOrdersMock(t)
	GetOrdersByUserEmailMock.EXPECT().GetOrdersByUserEmail(mock.Anything, "test@example.com").Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/orders/user/?email=test@example.com", nil)
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), GetOrdersByUserEmailMock)
	handler.GetOrdersByUserEmail(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetOrderByID_Success(t *testing.T) {
	GetOrderByIDMock := NewOrdersMock(t)
	GetOrderByIDMock.EXPECT().GetOrderByID(mock.Anything, 1).Return(&models.Order{ID: 1, UserEmail: "test@example.com"}, nil)
	GetOrderByIDMock.EXPECT().GetOrderItemsByOrderID(mock.Anything, 1).Return([]models.OrderItem{{ID: 1, OrderID: 1, ProductID: 1, Quantity: 2}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), GetOrderByIDMock)
	handler.GetOrderByID(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetOrderByID_BadRequest(t *testing.T) {
	GetOrderByIDMock := NewOrdersMock(t)
	req := httptest.NewRequest(http.MethodGet, "/orders/", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), GetOrderByIDMock)
	handler.GetOrderByID(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetOrderByID_NotFound(t *testing.T) {
	GetOrderByIDMock := NewOrdersMock(t)
	GetOrderByIDMock.EXPECT().GetOrderByID(mock.Anything, 1).Return(nil, storage.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), GetOrderByIDMock)
	handler.GetOrderByID(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetOrderByID_Fail(t *testing.T) {
	GetOrderByIDMock := NewOrdersMock(t)
	GetOrderByIDMock.EXPECT().GetOrderByID(mock.Anything, 1).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), GetOrderByIDMock)
	handler.GetOrderByID(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetOrderByID_FailOrderItems(t *testing.T) {
	GetOrderByIDMock := NewOrdersMock(t)
	GetOrderByIDMock.EXPECT().GetOrderByID(mock.Anything, 1).Return(&models.Order{ID: 1, UserEmail: "test@example.com", TotalPrice: 100.0}, nil)
	GetOrderByIDMock.EXPECT().GetOrderItemsByOrderID(mock.Anything, 1).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), GetOrderByIDMock)
	handler.GetOrderByID(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPlaceOrder_Success(t *testing.T) {
	items := []models.OrderItem{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 1},
	}

	jsonData, err := json.Marshal(CheckoutRequest{
		UserEmail: "test@example.com",
		Items:     items,
	})
	if err != nil {
		t.Fatal(err)
	}

	PlaceOrderMock := NewOrdersMock(t)
	PlaceOrderMock.EXPECT().PlaceOrder(mock.Anything, "test@example.com", items).Return(1, nil)

	req := httptest.NewRequest(http.MethodPost, "/orders/checkout", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), PlaceOrderMock)
	handler.PlaceOrder(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestPlaceOrder_BadRequest(t *testing.T) {
	PlaceOrderMock := NewOrdersMock(t)

	req := httptest.NewRequest(http.MethodPost, "/orders/checkout", bytes.NewReader([]byte(`invalid json`)))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), PlaceOrderMock)
	handler.PlaceOrder(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPlaceOrder_NotFound(t *testing.T) {
	items := []models.OrderItem{{ProductID: 999, Quantity: 1}}

	jsonData, err := json.Marshal(CheckoutRequest{
		UserEmail: "test@example.com",
		Items:     items,
	})
	if err != nil {
		t.Fatal(err)
	}

	PlaceOrderMock := NewOrdersMock(t)
	PlaceOrderMock.EXPECT().PlaceOrder(mock.Anything, "test@example.com", mock.Anything).Return(0, storage.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/orders/checkout", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), PlaceOrderMock)
	handler.PlaceOrder(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestPlaceOrder_Conflict(t *testing.T) {
	items := []models.OrderItem{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 1},
	}

	jsonData, err := json.Marshal(CheckoutRequest{
		UserEmail: "test@example.com",
		Items:     items,
	})
	if err != nil {
		t.Fatal(err)
	}

	PlaceOrderMock := NewOrdersMock(t)
	PlaceOrderMock.EXPECT().PlaceOrder(mock.Anything, "test@example.com", mock.Anything).Return(0, storage.ErrInsufficientStock)

	req := httptest.NewRequest(http.MethodPost, "/orders/checkout", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), PlaceOrderMock)
	handler.PlaceOrder(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestPlaceOrder_Fail(t *testing.T) {
	items := []models.OrderItem{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 1},
	}

	jsonData, err := json.Marshal(CheckoutRequest{
		UserEmail: "test@example.com",
		Items:     items,
	})
	if err != nil {
		t.Fatal(err)
	}

	PlaceOrderMock := NewOrdersMock(t)
	PlaceOrderMock.EXPECT().PlaceOrder(mock.Anything, "test@example.com", mock.Anything).Return(0, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/orders/checkout", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := NewHandler(slog.Default(), PlaceOrderMock)
	handler.PlaceOrder(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}
