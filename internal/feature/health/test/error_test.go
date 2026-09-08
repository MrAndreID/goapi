package health_test

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/MrAndreID/goapi/v2/internal/feature/health"
)

// Health hanya punya satu sentinel, jadi tidak ada tabel errorStatuses atau
// StatusFor untuk diuji. Yang penting dijaga adalah string sentinel-nya, karena
// itu bagian dari kontrak API yang dikirim di field Error saat probe gagal.
func TestErrDatabaseUnavailableCarriesTheContractString(t *testing.T) {
	if got := ErrDatabaseUnavailable.Error(); got != "DATABASE_UNAVAILABLE" {
		t.Fatalf("expected sentinel string %q, got %q", "DATABASE_UNAVAILABLE", got)
	}
}

// Sentinel harus tetap dikenali errors.Is meski dibungkus dengan %w, karena
// service membungkusnya saat melintasi layer.
func TestErrDatabaseUnavailableStaysRecognisableWhenWrapped(t *testing.T) {
	wrapped := fmt.Errorf("service layer: %w", ErrDatabaseUnavailable)

	if !errors.Is(wrapped, ErrDatabaseUnavailable) {
		t.Fatalf("expected wrapped error to still match ErrDatabaseUnavailable")
	}
}
