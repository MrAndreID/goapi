package health

import "github.com/MrAndreID/goapi/v2/internal/application/dependency"

const FeatureName string = "health"

func Dependency() dependency.Feature {
	return dependency.Feature{
		Name: FeatureName,
		Requirements: []dependency.Requirement{
			dependency.Database,
		},
	}
}
