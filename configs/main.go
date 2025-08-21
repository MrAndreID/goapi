package configs

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	AppName     string `env:"APP_NAME" envDefault:"Go Application Programming Interface (API)"`
	AppPort     string `env:"APP_PORT,notEmpty"`
	AppLocation string `env:"APP_LOCATION" envDefault:"Asia/Jakarta"`
	AppDebug    bool   `env:"APP_DEBUG" envDefault:"false"`
	AppVersion  string `env:"APP_VERSION" envDefault:"v1.0.0"`
	AppKey      string `env:"APP_KEY"`

	UseBodyDumpLog bool `env:"USE_BODY_DUMP_LOG" envDefault:"false"`

	UseDatabase        bool   `env:"USE_DATABASE" envDefault:"false"`
	DatabaseConnection string `env:"DATABASE_CONNECTION"`
	DatabaseHost       string `env:"DATABASE_HOST"`
	DatabasePort       string `env:"DATABASE_PORT"`
	DatabaseUsername   string `env:"DATABASE_USERNAME"`
	DatabasePassword   string `env:"DATABASE_PASSWORD"`
	DatabaseName       string `env:"DATABASE_NAME"`
	DatabaseSSLMode    string `env:"DATABASE_SSL_MODE" envDefault:"disable"`
	DatabaseParseTime  string `env:"DATABASE_PARSE_TIME" envDefault:"True"`
	DatabaseCharset    string `env:"DATABASE_CHARSET" envDefault:"utf8mb4"`
	DatabaseTimezone   string `env:"DATABASE_TIMEZONE" envDefault:"Asia/Jakarta"`

	UseCache        bool   `env:"USE_CACHE" envDefault:"false"`
	CacheConnection string `env:"CACHE_CONNECTION"`
	CacheHost       string `env:"CACHE_HOST"`
	CachePort       string `env:"CACHE_PORT"`
	CacheUsername   string `env:"CACHE_USERNAME"`
	CachePassword   string `env:"CACHE_PASSWORD"`

	UseObjectStorage        bool   `env:"USE_OBJECT_STORAGE" envDefault:"false"`
	ObjectStorageConnection string `env:"OBJECT_STORAGE_CONNECTION"`
	ObjectStorageHost       string `env:"OBJECT_STORAGE_HOST"`
	ObjectStoragePort       string `env:"OBJECT_STORAGE_PORT"`
	ObjectStorageUsername   string `env:"OBJECT_STORAGE_USERNAME"`
	ObjectStoragePassword   string `env:"OBJECT_STORAGE_PASSWORD"`
	ObjectStorageSSL        bool   `env:"OBJECT_STORAGE_SSL"`

	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envSeparator:","`
}

func New(toggle bool) (*Config, error) {
	var (
		tag string = "Configs.Main.New."
		cfg Config
	)

	logrus.SetFormatter(&logrus.JSONFormatter{})

	if err := godotenv.Load(); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to load environment file")

		return &cfg, err
	}

	if err := env.Parse(&cfg); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to parse environment")

		return &cfg, err
	}

	if cfg.UseBodyDumpLog {
		if err := NewBodyDumpLog(); err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "03",
				"error": err.Error(),
			}).Error("failed to initiate a body dump for log")

			return &cfg, err
		}
	}

	LoadVersion(&cfg, toggle)

	return &cfg, nil
}
