package application

import (
	"net/http"

	"github.com/MrAndreID/goapi/v2/internal/application/config"
	"github.com/MrAndreID/gomiddleware/v2"
	"github.com/MrAndreID/gopackage/v2"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/secure"
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
	e.Use(middleware.RequestLogger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowHeaders: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
	}))

	e.Use(gomiddleware.EchoSetNoCache)
	e.Use(gomiddleware.EchoSetMaintenanceMode("storage/maintenance.flag"))
	e.Use(gomiddleware.EchoCheckApplicationKey(cfg.AppKey))

	return e
}
