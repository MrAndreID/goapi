package user_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrAndreID/goapi/v2/internal/entity"
	. "github.com/MrAndreID/goapi/v2/internal/feature/v1/user"

	"github.com/labstack/echo/v5"
)

type stubService struct {
	createFunc func(CreateData) (User, error)
	readFunc   func(context.Context, ReadData) (entity.PaginatorResponse, error)
	updateFunc func(UpdateData) error
	deleteFunc func(DeleteData) error
}

func (s *stubService) Create(ctx context.Context, req CreateData) (User, error) {
	if s.createFunc != nil {
		return s.createFunc(req)
	}
	return User{}, nil
}

func (s *stubService) Read(ctx context.Context, req ReadData) (entity.PaginatorResponse, error) {
	if s.readFunc != nil {
		return s.readFunc(ctx, req)
	}
	return entity.PaginatorResponse{}, nil
}

func (s *stubService) Update(ctx context.Context, req UpdateData) error {
	if s.updateFunc != nil {
		return s.updateFunc(req)
	}
	return nil
}

func (s *stubService) Delete(ctx context.Context, req DeleteData) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(req)
	}
	return nil
}

func TestHandlerCreateReturnsBadRequestOnValidationError(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	NewHandler(group, &stubService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", strings.NewReader(`{"emails":["andre@gmail.com"]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandlerCreateReturnsCreated(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	service := &stubService{createFunc: func(req CreateData) (User, error) {
		return User{ID: "user-1", Name: req.Name, Emails: []Email{{Email: req.Emails[0]}}}, nil
	}}
	NewHandler(group, service)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", strings.NewReader(`{"name":"Andre","emails":["andre@gmail.com"]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response entity.MainResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Code != "0201" {
		t.Fatalf("expected code 0201, got %s", response.Code)
	}
}

func TestHandlerReadReturnsSuccess(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	service := &stubService{readFunc: func(ctx context.Context, req ReadData) (entity.PaginatorResponse, error) {
		return entity.PaginatorResponse{Records: []User{{ID: "user-1"}}, Total: 1}, nil
	}}
	NewHandler(group, service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user", nil)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandlerUpdateReturnsSuccess(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	service := &stubService{updateFunc: func(req UpdateData) error { return nil }}
	NewHandler(group, service)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/user/123e4567-e89b-12d3-a456-426614174000", strings.NewReader(`{"name":"Andre"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandlerDeleteReturnsSuccess(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	service := &stubService{deleteFunc: func(req DeleteData) error { return nil }}
	NewHandler(group, service)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user/123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandlerReadReturnsInternalServerError(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	service := &stubService{readFunc: func(ctx context.Context, req ReadData) (entity.PaginatorResponse, error) {
		return entity.PaginatorResponse{}, errors.New("boom")
	}}
	NewHandler(group, service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user", nil)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandlerUpdateReturnsInternalServerError(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	service := &stubService{updateFunc: func(req UpdateData) error { return errors.New("boom") }}
	NewHandler(group, service)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/user/123e4567-e89b-12d3-a456-426614174000", strings.NewReader(`{"name":"Andre"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandlerDeleteReturnsInternalServerError(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	service := &stubService{deleteFunc: func(req DeleteData) error { return errors.New("boom") }}
	NewHandler(group, service)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user/123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandlerCreateReturnsInternalServerError(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	service := &stubService{createFunc: func(req CreateData) (User, error) {
		return User{}, errors.New("boom")
	}}
	NewHandler(group, service)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", strings.NewReader(`{"name":"Andre","emails":["andre@gmail.com"]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandlerReadReturnsBadRequestOnValidationError(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	NewHandler(group, &stubService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user?id=not-a-uuid", nil)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandlerUpdateReturnsBadRequestOnValidationError(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	NewHandler(group, &stubService{})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/user/not-a-uuid", strings.NewReader(`{"name":"Andre"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandlerDeleteReturnsBadRequestOnValidationError(t *testing.T) {
	echoServer := echo.New()
	group := echoServer.Group("/api/v1")
	NewHandler(group, &stubService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	echoServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
