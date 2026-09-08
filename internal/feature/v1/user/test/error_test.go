package user_test

import (
	"fmt"
	"net/http"
	"testing"

	. "github.com/MrAndreID/goapi/v2/internal/feature/v1/user"
)

func TestStatusForMapsKnownSentinels(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected int
	}{
		{"duplicate email is conflict", ErrDuplicateEmail, http.StatusConflict},
		{"missing user is not found", ErrFailedToReadUserData, http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := StatusFor(tc.err); got != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, got)
			}
		})
	}
}

func TestStatusForFallsBackToInternalServerError(t *testing.T) {
	if got := StatusFor(ErrFailedToCreateUser); got != http.StatusInternalServerError {
		t.Fatalf("expected %d for an unmapped sentinel, got %d", http.StatusInternalServerError, got)
	}
}

func TestStatusForMatchesWrappedError(t *testing.T) {
	wrapped := fmt.Errorf("service layer: %w", ErrDuplicateEmail)

	if got := StatusFor(wrapped); got != http.StatusConflict {
		t.Fatalf("expected wrapped sentinel to stay recognisable as %d, got %d", http.StatusConflict, got)
	}
}
