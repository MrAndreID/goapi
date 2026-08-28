package user_test

import (
	"testing"

	"github.com/MrAndreID/goapi/v2/internal/application/dependency"
	. "github.com/MrAndreID/goapi/v2/internal/feature/v1/user"
)

func TestDependencyRequiresDatabase(t *testing.T) {
	feature := Dependency()

	if feature.Name != FeatureName {
		t.Fatalf("expected name %q, got %q", FeatureName, feature.Name)
	}

	var found bool

	for _, requirement := range feature.Requirements {
		if requirement == dependency.Database {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected %q to require %q", feature.Name, dependency.Database)
	}
}
