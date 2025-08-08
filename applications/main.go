package applications

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MrAndreID/goapi/caches"
	"github.com/MrAndreID/goapi/configs"
	"github.com/MrAndreID/goapi/databases"
	"github.com/MrAndreID/goapi/internal/handlers"

	"github.com/MrAndreID/gomiddleware"
	"github.com/MrAndreID/gopackage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/secure"
	"go.elastic.co/apm/module/apmechov4"
	"gorm.io/gorm"
)

type Application struct {
	Config       *configs.Config
	TimeLocation *time.Location
	Database     *gorm.DB
	Cache        *redis.Client
}

func Start(toggle bool) any {
	var tag string = "Applications.Main.New."

	cfg, err := configs.New(toggle)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to initiate configuration")

		return nil
	}

	timeLocation, err := time.LoadLocation(cfg.AppLocation)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to load location for time")

		return nil
	}

	var databaseConnection *gorm.DB

	if cfg.UseDatabase {
		databaseConnection, err = databases.New(&databases.Database{
			Connection: cfg.DatabaseConnection,
			Host:       cfg.DatabaseHost,
			Port:       cfg.DatabasePort,
			Username:   cfg.DatabaseUsername,
			Password:   cfg.DatabasePassword,
			Name:       cfg.DatabaseName,
			SSLMode:    cfg.DatabaseSSLMode,
			ParseTime:  cfg.DatabaseParseTime,
			Charset:    cfg.DatabaseCharset,
			Timezone:   cfg.DatabaseTimezone,
		}, cfg.AppDebug)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "03",
				"error": err.Error(),
			}).Error("failed to connect database")

			return nil
		}
	}

	var cacheConnection *redis.Client

	if cfg.UseCache {
		cacheConnection, err = caches.New(&caches.Cache{
			Connection: cfg.CacheConnection,
			Host:       cfg.CacheHost,
			Port:       cfg.CachePort,
			Username:   cfg.CacheUsername,
			Password:   cfg.CachePassword,
		})

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "04",
				"error": err.Error(),
			}).Error("failed to connect cache")

			return nil
		}
	}

	app := Application{
		Config:       cfg,
		TimeLocation: timeLocation,
		Database:     databaseConnection,
		Cache:        cacheConnection,
	}

	echo.NotFoundHandler = func(c echo.Context) error {
		logrus.WithFields(logrus.Fields{
			"tag": tag + "01",
		}).Error("route not found")

		return c.JSON(http.StatusNotFound, map[string]string{
			"code":        fmt.Sprintf("%04d", http.StatusNotFound),
			"description": strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusNotFound), " ", "_")),
		})
	}

	echo.MethodNotAllowedHandler = func(c echo.Context) error {
		logrus.WithFields(logrus.Fields{
			"tag": tag + "02",
		}).Error("method not allowed")

		return c.JSON(http.StatusMethodNotAllowed, map[string]string{
			"code":        fmt.Sprintf("%04d", http.StatusMethodNotAllowed),
			"description": strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusMethodNotAllowed), " ", "_")),
		})
	}

	e := echo.New()

	e.Validator = gopackage.CustomValidator()

	e.HTTPErrorHandler = gopackage.EchoCustomHTTPErrorHandler

	e.JSONSerializer = gopackage.CustomJSON()

	e.Pre(middleware.RemoveTrailingSlash())

	e.Pre(gomiddleware.EchoSetRequestID)

	e.Use(apmechov4.Middleware())

	e.Use(middleware.Recover())

	e.Use(middleware.BodyDump(func(c echo.Context, requestBody, responseBody []byte) {
		request := struct {
			Header interface{} `json:"header"`
			Body   string      `json:"body"`
		}{
			Header: c.Request().Header,
			Body:   string(requestBody),
		}

		response := struct {
			Header interface{} `json:"header"`
			Body   string      `json:"body"`
		}{
			Header: c.Response().Header(),
			Body:   string(responseBody),
		}

		logrus.WithFields(logrus.Fields{
			"request":   request,
			"requestId": c.Get("RequestID"),
			"response":  response,
			"url":       c.Request().Host + c.Request().URL.String(),
		}).Info("body dump")
	}))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	secureMiddleware := secure.Options{
		SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:           63072000,
		STSIncludeSubdomains: true,
		STSPreload:           true,
		ForceSTSHeader:       true,
		IsDevelopment:        true,
	}

	e.Use(echo.WrapMiddleware(secure.New(secureMiddleware).Handler))

	e.Use(middleware.Logger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowHeaders: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	e.Use(gomiddleware.EchoSetNoCache)

	e.Use(gomiddleware.EchoSetMaintenanceMode("storages/maintenance.flag"))

	if cfg.AppDebug {
		e.Logger.SetLevel(log.DEBUG)

		e.Debug = true
	}

	initService(&app)

	api := e.Group("/api")

	v1 := api.Group("/v1")

	if toggle {
		handlers.NewUserHandler(v1, UserService)

		return e.Start(":" + cfg.AppPort)
	}

	return v1
}
