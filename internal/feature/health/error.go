package health

import (
	"errors"
)

var (
	ErrServiceUnhealthy         = errors.New("SERVICE_UNHEALTHY")
	ErrDatabaseUnavailable      = errors.New("DATABASE_UNAVAILABLE")
	ErrCacheUnavailable         = errors.New("CACHE_UNAVAILABLE")
	ErrMessageBrokerUnavailable = errors.New("MESSAGE_BROKER_UNAVAILABLE")
	ErrObjectStorageUnavailable = errors.New("OBJECT_STORAGE_UNAVAILABLE")
)
