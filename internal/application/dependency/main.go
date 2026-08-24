package dependency

import (
	"errors"
	"strings"
)

type Requirement string

type Availability map[Requirement]bool

type Feature struct {
	Name         string
	Requirements []Requirement
}

const (
	Database      Requirement = "database"
	Cache         Requirement = "cache"
	MessageBroker Requirement = "message broker"
	ObjectStorage Requirement = "object storage"
)

func (requirement Requirement) EnvironmentKey() string {
	switch requirement {
	case Database:
		return "USE_DATABASE"
	case Cache:
		return "USE_CACHE"
	case MessageBroker:
		return "USE_MESSAGE_BROKER"
	case ObjectStorage:
		return "USE_OBJECT_STORAGE"
	default:
		return ""
	}
}

func Validate(features []Feature, availability Availability) error {
	return validate(features, availability, func(requirement Requirement) string {
		return "set " + requirement.EnvironmentKey() + "=true and fill its configuration in the environment file"
	})
}

func validate(features []Feature, availability Availability, hint func(Requirement) string) error {
	var problems []string

	for _, feature := range features {
		for _, requirement := range feature.Requirements {
			if requirement.EnvironmentKey() == "" {
				problems = append(problems, "feature "+feature.Name+" declares an unknown requirement "+string(requirement))

				continue
			}

			if !availability[requirement] {
				problems = append(problems, "feature "+feature.Name+" needs "+string(requirement)+", "+hint(requirement))
			}
		}
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}

	return nil
}
