package product

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Get Product - Ready
func TestGetAllProducts_Success(t *testing.T) {
	// Мокаем storage — он вернёт один продукт.
	GetAllProductsmock := NewProductsMock(t)
	GetAllProductsmock.EXPECT().GetAllProducts(mock.Anything).Return([]models.Product{
		{ID: 1, Name: "Dog Food"},
	}, nil)

	// Создаем HTTP-запрос GET /products
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	// Создаем хендлер с мок-хранилищем
	handler := New(slog.Default(), GetAllProductsmock)

	// Вызываем метод GetAllProducts, который является http.HandlerFunc
	handler.GetAllProducts(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
func TestGetAllProducts_Error(t *testing.T) {
	// Мокаем storage — он будет возвращать ошибку
	GetAllProductsmock := NewProductsMock(t)
	GetAllProductsmock.EXPECT().GetAllProducts(mock.Anything).Return(nil, errors.New("DB error"))

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), GetAllProductsmock)
	handler.GetAllProducts(w, req)

	// Ожидаем HTTP 500
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetProductByID_Success(t *testing.T) {
	// Мокаем storage — он вернёт продукт с ID 1.
	GetProductByIDmock := NewProductsMock(t)
	GetProductByIDmock.EXPECT().GetProductByID(mock.Anything, 1).Return(&models.Product{ID: 1, Name: "Dog Food"}, nil)

	// Создаем HTTP-запрос GET /products/1
	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	// Создаем хендлер с мок-хранилищем
	handler := New(slog.Default(), GetProductByIDmock)

	// Вызываем метод GetProductByID
	handler.GetProductByID(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetProductByID_NotFound(t *testing.T) {
	// Мокаем storage — он вернёт ошибку "not found".
	GetProductByIDmock := NewProductsMock(t)
	GetProductByIDmock.EXPECT().GetProductByID(mock.Anything, 999).Return(nil, storage.ErrNotFound)

	// Создаем HTTP-запрос GET /products/999
	req := httptest.NewRequest(http.MethodGet, "/products/999", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	// Создаем хендлер с мок-хранилищем
	handler := New(slog.Default(), GetProductByIDmock)

	// Вызываем метод GetProductByID
	handler.GetProductByID(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetProductByID_BadRequest(t *testing.T) {
	// Создаем HTTP-запрос GET /products/abc (невалидный ID)
	GetProductByIDmock := NewProductsMock(t)

	req := httptest.NewRequest(http.MethodGet, "/products/abc", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	// Создаем хендлер с мок-хранилищем
	handler := New(slog.Default(), GetProductByIDmock)

	// Вызываем метод GetProductByID
	handler.GetProductByID(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetProductByID_Fail(t *testing.T) {
	// Мокаем storage — он вернёт ошибку сервиса.
	GetProductByIDmock := NewProductsMock(t)
	GetProductByIDmock.EXPECT().GetProductByID(mock.Anything, 1).Return(nil, errors.New("DB error"))

	// Создаем HTTP-запрос GET /products/1
	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	// Создаем хендлер с мок-хранилищем
	handler := New(slog.Default(), GetProductByIDmock)

	// Вызываем метод GetProductByID
	handler.GetProductByID(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// =======================
// Create Product
// =======================

func TestCreateProduct_Success(t *testing.T) {
	// TODO: Написать Unit-тест для создания продукта (200 OK)
	newProduct := models.Product{
		ID:    2,
		Name:  "Cat's Food",
		Price: 10.5,
		Stock: 500,
	}

	jsonData, err := json.Marshal(newProduct)
	if err != nil {
		t.Fatal(err)
	}

	CreateProductsmock := NewProductsMock(t)
	CreateProductsmock.EXPECT().CreateProduct(mock.Anything, newProduct).Return(newProduct.ID, nil)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateProductsmock)
	handler.CreateProduct(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateProduct_BadRequest(t *testing.T) {
	// TODO: Написать Unit-тест для создания продукта с невалидным JSON (400 Bad Request)
	jsonData := []byte(`{
		id: 3,
		"name": "Dogs",
		"price": 2.5,
		"stock": 123
	`)

	CreateProductsmock := NewProductsMock(t)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateProductsmock)
	handler.CreateProduct(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateProduct_Fail(t *testing.T) {
	// TODO: Написать Unit-тест для создания продукта при ошибке сервиса (500 Internal Server Error)
	newProduct := models.Product{
		ID:    3,
		Name:  "Dog's Toy",
		Price: 2.5,
		Stock: 123,
	}

	jsonData, _ := json.Marshal(newProduct)

	CreateProductsmock := NewProductsMock(t)
	CreateProductsmock.EXPECT().CreateProduct(mock.Anything, newProduct).Return(0, errors.New("DB error"))

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateProductsmock)
	handler.CreateProduct(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// =======================
// Update Product
// =======================

func TestUpdateProduct_Success(t *testing.T) {
	// TODO: Написать Unit-тест для обновления продукта (200 OK)
	updateProduct := models.Product{
		ID:    4,
		Name:  "Rat's Food",
		Price: 32.1,
		Stock: 1000,
	}

	jsonData, err := json.Marshal(updateProduct)
	if err != nil {
		t.Fatal(err)
	}

	UpdateProductsmock := NewProductsMock(t)
	UpdateProductsmock.EXPECT().UpdateProduct(mock.Anything, updateProduct).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/products/4", bytes.NewReader(jsonData))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "4")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler := New(slog.Default(), UpdateProductsmock)
	handler.UpdateProduct(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateProduct_BadRequest(t *testing.T) {
	// TODO: Написать Unit-тест для обновления продукта с невалидным JSON (400 Bad Request)
	jsonData := []byte(`{
		id: 4,
		"name": "Rat's Food",
		price: 32.1,
		"stock": 200
	`)

	UpdateProductsmock := NewProductsMock(t)

	req := httptest.NewRequest(http.MethodPut, "/products/4", bytes.NewReader(jsonData))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "4")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler := New(slog.Default(), UpdateProductsmock)
	handler.UpdateProduct(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProduct_Fail(t *testing.T) {
	// TODO: Написать Unit-тест для обновления продукта при ошибке сервиса (500 Internal Server Error)
	updateProduct := models.Product{
		ID:    4,
		Name:  "Rat's Food",
		Price: 32.1,
		Stock: 1000,
	}

	jsonData, err := json.Marshal(updateProduct)
	if err != nil {
		t.Fatal(err)
	}

	UpdateProductsmock := NewProductsMock(t)
	UpdateProductsmock.EXPECT().UpdateProduct(mock.Anything, updateProduct).Return(errors.New("DB error"))

	req := httptest.NewRequest(http.MethodPut, "/products/4", bytes.NewReader(jsonData))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "4")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler := New(slog.Default(), UpdateProductsmock)
	handler.UpdateProduct(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// =======================
// Delete Product
// =======================

func TestDeleteProduct_Success(t *testing.T) {
	// TODO: Написать Unit-тест для удаления продукта (200 OK)
	DeleteProductsmock := NewProductsMock(t)
	DeleteProductsmock.EXPECT().DeleteProduct(mock.Anything, 4).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/products/4", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "4")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler := New(slog.Default(), DeleteProductsmock)
	handler.DeleteProduct(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteProduct_BadRequest(t *testing.T) {
	// TODO: Написать Unit-тест для удаления продукта с невалидным ID (400 Bad Request)
	DeleteProductsmock := NewProductsMock(t)

	req := httptest.NewRequest(http.MethodDelete, "/products", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler := New(slog.Default(), DeleteProductsmock)
	handler.DeleteProduct(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteProduct_Fail(t *testing.T) {
	// TODO: Написать Unit-тест для удаления продукта при ошибке сервиса (500 Internal Server Error)
	DeleteProductsmock := NewProductsMock(t)
	DeleteProductsmock.EXPECT().DeleteProduct(mock.Anything, 4).Return(errors.New("DB error"))

	req := httptest.NewRequest(http.MethodDelete, "/products/4", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "4")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler := New(slog.Default(), DeleteProductsmock)
	handler.DeleteProduct(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetPopularProducts_Success(t *testing.T) {
	popular := []models.PopularProduct{
		{ProductID: 1, Name: "Dog Food", TotalSold: 10},
		{ProductID: 2, Name: "Cat Food", TotalSold: 7},
	}

	GetPopularProductsmock := NewProductsMock(t)
	GetPopularProductsmock.EXPECT().GetPopularProducts(mock.Anything).Return(popular, nil)

	req := httptest.NewRequest(http.MethodGet, "/products/popular", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), GetPopularProductsmock)
	handler.GetPopularProducts(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetPopularProducts_Fail(t *testing.T) {
	GetPopularProductsmock := NewProductsMock(t)
	GetPopularProductsmock.EXPECT().GetPopularProducts(mock.Anything).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/products/popular", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), GetPopularProductsmock)
	handler.GetPopularProducts(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}
