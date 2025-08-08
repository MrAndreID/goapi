package tests

import (
	"encoding/json"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type (
	Request struct {
		Method    string
		Url       string
		PathParam *PathParam
	}

	PathParam struct {
		Name  string
		Value string
	}

	TestCase struct {
		TestName      string
		Request       Request
		RequestHeader *[]Header
		RequestBody   any
		HandlerFunc   func(c echo.Context) error
		Expected      ExpectedResponse
	}

	Header struct {
		Key   string
		Value string
	}

	ExpectedResponse struct {
		StatusCode int
		BodyPart   Response
	}

	Response struct {
		Code        string `json:"code"`
		Description string `json:"description"`
		Data        any    `json:"data"`
	}
)

func PrepareContextFromTestCase(test TestCase) (c echo.Context, recorder *httptest.ResponseRecorder) {
	requestJson, err := json.Marshal(test.RequestBody)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "Tests.Main.PrepareContextFromTestCase.01",
			"error": err.Error(),
		}).Error("failed to json marshal from request body")
	}

	request := httptest.NewRequest(test.Request.Method, test.Request.Url, strings.NewReader(string(requestJson)))

	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	if test.RequestHeader != nil {
		for _, v := range *test.RequestHeader {
			request.Header.Set(v.Key, v.Value)
		}
	}

	recorder = httptest.NewRecorder()

	e := echo.New()

	c = e.NewContext(request, recorder)

	if test.Request.PathParam != nil {
		c.SetParamNames(test.Request.PathParam.Name)

		c.SetParamValues(test.Request.PathParam.Value)
	}

	return c, recorder
}
