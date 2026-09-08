package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrAndreID/goapi/v2/internal/entity"
	. "github.com/MrAndreID/goapi/v2/internal/feature/health"

	"github.com/labstack/echo/v5"
)

type stubService struct {
	checkFunc func(context.Context) error
}

func (s *stubService) Check(ctx context.Context) error {
	if s.checkFunc != nil {
		return s.checkFunc(ctx)
	}
	return nil
}

func TestHandlerCheckReturnsOKWhenHealthy(t *testing.T) {
	e := echo.New()
	NewHandler(e, &stubService{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response entity.MainResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.Status {
		t.Fatalf("expected status true, got %t", response.Status)
	}

	if response.Error != nil {
		t.Fatalf("expected no error detail on a healthy probe, got %v", response.Error)
	}
}

func TestHandlerCheckReturnsServiceUnavailableWhenDatabaseDown(t *testing.T) {
	e := echo.New()
	NewHandler(e, &stubService{checkFunc: func(ctx context.Context) error {
		return ErrDatabaseUnavailable
	}})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var response entity.MainResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status {
		t.Fatalf("expected status false on an unhealthy probe")
	}

	details, ok := response.Error.([]any)
	if !ok || len(details) != 1 || details[0] != ErrServiceUnhealthy.Error() {
		t.Fatalf("expected the sentinel detail to be attached, got %v", response.Error)
	}
}

// Health tidak membedakan jenis kegagalan: probe apa pun yang gagal, termasuk
// error tak terduga dari service, dipetakan ke satu status 503 dengan detail
// sentinel yang sama. Tidak ada cabang 500 seperti pada feature bisnis.
func TestHandlerCheckReturnsServiceUnavailableOnAnyError(t *testing.T) {
	e := echo.New()
	NewHandler(e, &stubService{checkFunc: func(ctx context.Context) error {
		return errors.New("unexpected failure")
	}})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var response entity.MainResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	details, ok := response.Error.([]any)
	if !ok || len(details) != 1 || details[0] != ErrServiceUnhealthy.Error() {
		t.Fatalf("expected the sentinel detail to be attached, got %v", response.Error)
	}
}
