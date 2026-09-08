package version_test

import (
	"context"
	"testing"

	. "github.com/MrAndreID/goapi/v2/internal/feature/version"
)

func TestServiceReadReturnsNameAndVersion(t *testing.T) {
	svc := NewService("GoAPI", "v1.2.3")

	info := svc.Read(context.Background())

	if info.Name != "GoAPI" || info.Version != "v1.2.3" {
		t.Fatalf("unexpected version payload: %+v", info)
	}
}
