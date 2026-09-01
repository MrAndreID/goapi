package application

import (
	"errors"
	"net/http"
	"time"

	"github.com/MrAndreID/goapi/v2/internal/application/config"

	"github.com/MrAndreID/gomiddleware/v2"
	"github.com/MrAndreID/gopackage/v2"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/sirupsen/logrus"
)

func newServer(cfg *config.Config) *echo.Echo {
	e := echo.New()

	e.Validator = gopackage.CustomValidator()
	e.HTTPErrorHandler = gopackage.EchoCustomHTTPErrorHandler
	e.JSONSerializer = gopackage.CustomJSONSerializer()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(gomiddleware.EchoSetRequestID)

	e.Use(middleware.Recover())

	e.Use(middleware.BodyDump(func(c *echo.Context, requestBody, responseBody []byte, err error) {
		request := struct {
			Header any    `json:"header"`
			Body   string `json:"body"`
		}{
			Header: c.Request().Header,
			Body:   string(requestBody),
		}

		response := struct {
			Header any    `json:"header"`
			Body   string `json:"body"`
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

	secureMiddleware := middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            63072000,
		HSTSPreloadEnabled:    true,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}

	if cfg.AppDebug {
		secureMiddleware.HSTSMaxAge = 0
		secureMiddleware.HSTSPreloadEnabled = false
	}

	e.Use(middleware.SecureWithConfig(secureMiddleware))
	e.Use(middleware.RequestLogger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowHeaders: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
	}))

	e.Use(gomiddleware.EchoSetNoCache)
	e.Use(gomiddleware.EchoSetMaintenanceMode("storage/maintenance.flag"))

	e.Use(middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
		Timeout: time.Duration(cfg.AppTimeout) * time.Second,
		ErrorHandler: func(c *echo.Context, err error) error {
			return errors.New("REQUEST_TIMEOUT")
		},
	}))

	rateLimiterConfig := middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: cfg.AppRateLimit, Burst: int(cfg.AppRateLimit), ExpiresIn: time.Duration(cfg.AppRateLimitDeadline) * time.Second},
		),
		ErrorHandler: func(c *echo.Context, err error) error {
			return errors.New("FORBIDDEN")
		},
		DenyHandler: func(c *echo.Context, identifier string, err error) error {
			return errors.New("TOO_MANY_REQUESTS")
		},
	}

	e.Use(middleware.RateLimiterWithConfig(rateLimiterConfig))

	return e
}
