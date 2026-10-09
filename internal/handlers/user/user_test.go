package user

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

func TestCreateUser_Success(t *testing.T) {
	newUser := models.User{
		ID:    2,
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	jsonData, err := json.Marshal(newUser)
	if err != nil {
		t.Fatal(err)
	}

	CreateUsermock := NewUserMock(t)
	CreateUsermock.EXPECT().CreateUser(mock.Anything, newUser).Return(newUser.ID, nil)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateUsermock)
	handler.CreateUser(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateUser_BadRequest(t *testing.T) {
	newUser := []byte(`{"id": 2, "name": "John Doe", "email": "john.doe@example.com"`)

	CreateUsermock := NewUserMock(t)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(newUser))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateUsermock)
	handler.CreateUser(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUser_Fail(t *testing.T) {

	newUser := models.User{
		ID:    2,
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	CreateUsermock := NewUserMock(t)
	CreateUsermock.EXPECT().CreateUser(mock.Anything, newUser).Return(0, assert.AnError)

	jsonData, err := json.Marshal(newUser)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateUsermock)
	handler.CreateUser(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetUserByEmail_Success(t *testing.T) {
	CreateUsermock := NewUserMock(t)
	CreateUsermock.EXPECT().GetUserByEmail(mock.Anything, "john.doe@example.com").Return(models.User{
		ID:    2,
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/email/john.doe@example.com", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("email", "john.doe@example.com")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateUsermock)
	handler.GetUserByEmail(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserByEmail_BadRequest(t *testing.T) {
	CreateUsermock := NewUserMock(t)

	req := httptest.NewRequest(http.MethodGet, "/users/email/", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("email", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateUsermock)
	handler.GetUserByEmail(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	CreateUsermock := NewUserMock(t)
	CreateUsermock.EXPECT().GetUserByEmail(mock.Anything, "not.exist@example.com").Return(models.User{}, storage.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/users/email/not.exist@example.com", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("email", "not.exist@example.com")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateUsermock)
	handler.GetUserByEmail(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserByEmail_Fail(t *testing.T) {
	CreateUsermock := NewUserMock(t)
	CreateUsermock.EXPECT().GetUserByEmail(mock.Anything, "test@example.com").Return(models.User{}, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/users/email/test@example.com", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("email", "test@example.com")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateUsermock)
	handler.GetUserByEmail(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetAllUsers_Success(t *testing.T) {
	CreateUsermock := NewUserMock(t)
	CreateUsermock.EXPECT().GetAllUsers(mock.Anything).Return([]models.User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateUsermock)
	handler.GetAllUsers(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetAllUsers_Fail(t *testing.T) {
	CreateUsermock := NewUserMock(t)
	CreateUsermock.EXPECT().GetAllUsers(mock.Anything).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), CreateUsermock)
	handler.GetAllUsers(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}
