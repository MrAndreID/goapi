package version_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/MrAndreID/goapi/v2/internal/feature/version"

	"github.com/labstack/echo/v5"
)

type stubService struct {
	readFunc func(context.Context) Info
}

func (s *stubService) Read(ctx context.Context) Info {
	if s.readFunc != nil {
		return s.readFunc(ctx)
	}
	return Info{}
}

func TestHandlerReadReturnsNameAndVersion(t *testing.T) {
	e := echo.New()
	service := &stubService{readFunc: func(ctx context.Context) Info {
		return Info{Name: "GoAPI", Version: "v1.2.3"}
	}}
	NewHandler(e, service)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		Status bool `json:"status"`
		Data   Info `json:"data"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.Status {
		t.Fatalf("expected status true, got %t", response.Status)
	}

	if response.Data.Name != "GoAPI" || response.Data.Version != "v1.2.3" {
		t.Fatalf("unexpected version payload: %+v", response.Data)
	}
}
