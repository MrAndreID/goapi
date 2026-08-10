package application

import (
	"time"

	"github.com/MrAndreID/goapi/v2/internal/application/cache"
	"github.com/MrAndreID/goapi/v2/internal/application/config"
	messageBroker "github.com/MrAndreID/goapi/v2/internal/application/message_broker"
	objectStorage "github.com/MrAndreID/goapi/v2/internal/application/object_storage"

	"gorm.io/gorm"
)

type Application struct {
	Config        *config.Config
	TimeLocation  *time.Location
	Database      *gorm.DB
	Cache         *cache.CacheConnection
	ObjectStorage *objectStorage.ObjectStorageConnection
	MessageBroker *messageBroker.MessageBrokerConnection
}
