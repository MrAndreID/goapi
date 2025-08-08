package tests

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"testing"
	"time"

	"github.com/MrAndreID/goapi/applications"
	"github.com/MrAndreID/goapi/internal/handlers"
	"github.com/MrAndreID/goapi/internal/types"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var id string

var userHandlerFunc = handlers.NewUserHandler(applications.Start(false).(*echo.Group), applications.UserService)

func UserDataTest(t *testing.T, expectedData, data any) {
	recorderResponseDataBytes, err := json.Marshal(data)

	assert.Condition(t, func() bool {
		return err == nil
	}, "Failed to JSON Marshal for Recorder Response Data. Actual: %v", data)

	var recorderResponseData map[string]any

	json.Unmarshal(recorderResponseDataBytes, &recorderResponseData)

	assert.Condition(t, func() bool {
		val, ok := recorderResponseData["id"].(string)

		if !ok {
			return false
		}

		id = val

		_, err := uuid.Parse(val)

		return err == nil
	}, "Expected the ID in UUID form. Actual: %v", recorderResponseData["id"])

	assert.Condition(t, func() bool {
		val, ok := recorderResponseData["createdAt"].(string)

		if !ok {
			return false
		}

		_, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", val)

		return err == nil
	}, "Expected the Created At in datatime form (2006-01-02T15:04:05.999999999Z07:00). Actual: %v", recorderResponseData["createdAt"])

	assert.Condition(t, func() bool {
		val, ok := recorderResponseData["updatedAt"].(string)

		if !ok {
			return false
		}

		_, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", val)

		return err == nil
	}, "Expected the Updated At in datatime form (2006-01-02T15:04:05.999999999Z07:00). Actual: %v", recorderResponseData["updatedAt"])

	if expectedData != nil {
		assert.Equal(t, expectedData.(map[string]any)["deletedAt"], recorderResponseData["deletedAt"])

		assert.Condition(t, func() bool {
			if expectedData.(map[string]any)["name"] == nil {
				return true
			}

			return recorderResponseData["name"] == expectedData.(map[string]any)["name"]
		}, "Expected the Name in string form. Actual: %v", recorderResponseData["name"])
	}

	val, ok := recorderResponseData["emails"].([]any)

	assert.Condition(t, func() bool {
		if !ok {
			return false
		}

		return len(val) != 0
	}, "Expected the Emails is zero data.")

	emailRecorderResponseDataBytes, err := json.Marshal(val)

	assert.Condition(t, func() bool {
		return err == nil
	}, "Failed to JSON Marshal for Email Recorder Response Data. Actual: %v", val)

	var emailsRecorderResponseData []map[string]any

	json.Unmarshal(emailRecorderResponseDataBytes, &emailsRecorderResponseData)

	for i, v := range emailsRecorderResponseData {
		assert.Condition(t, func() bool {
			val, ok := v["id"].(string)

			if !ok {
				return false
			}

			_, err := uuid.Parse(val)

			return err == nil
		}, "Expected the ID from emails in UUID form. Actual: %v", v["id"])

		assert.Condition(t, func() bool {
			val, ok := v["createdAt"].(string)

			if !ok {
				return false
			}

			_, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", val)

			return err == nil
		}, "Expected the Created At from emails in datatime form (2006-01-02T15:04:05.999999999Z07:00). Actual: %v", v["createdAt"])

		assert.Condition(t, func() bool {
			val, ok := v["updatedAt"].(string)

			if !ok {
				return false
			}

			_, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", val)

			return err == nil
		}, "Expected the Updated At from emails in datatime form (2006-01-02T15:04:05.999999999Z07:00). Actual: %v", v["updatedAt"])

		if v["deletedAt"] != nil {
			assert.Condition(t, func() bool {
				val, ok := v["deletedAt"].(string)

				if !ok {
					return false
				}

				_, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", val)

				return err == nil
			}, "Expected the Deleted At from emails in datatime form (2006-01-02T15:04:05.999999999Z07:00). Actual: %v", v["deletedAt"])
		} else {
			if expectedData.(map[string]any)["emails"] != nil {
				if expectedData.(map[string]any)["emails"].([]map[string]any)[i]["deletedAt"] != nil {
					assert.Equal(t, expectedData.(map[string]any)["emails"].([]map[string]any)[i]["deletedAt"], v["deletedAt"])
				}
			}
		}

		assert.Equal(t, id, v["userId"])

		if expectedData.(map[string]any)["emails"] != nil {
			if expectedData.(map[string]any)["emails"].([]map[string]any)[i]["email"] != nil {
				assert.Equal(t, expectedData.(map[string]any)["emails"].([]map[string]any)[i]["email"], v["email"])
			}
		} else {
			assert.Condition(t, func() bool {
				val, ok := v["email"].(string)

				if !ok {
					return false
				}

				_, err := mail.ParseAddress(val)

				return err == nil
			}, "Expected the Email from emails in email form. Actual: %v", v["email"])
		}
	}
}

func TestCreateUser(t *testing.T) {
	headers := []Header{
		{
			Key:   "Content-Type",
			Value: "application/json",
		},
	}

	var handlerFunc = func(c echo.Context) error {
		return userHandlerFunc.Create(c)
	}

	cases := []TestCase{
		{
			"Create User => Failed Validation => Name (Required)",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "cannot be blank",
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Name (Blacklist => ')",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test'",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			`Create User => Failed Validation => Name (Blacklist => ")`,
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: `Unit Test"`,
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Name (Blacklist => [)",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test[",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Name (Blacklist => ])",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test]",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Name (Blacklist => <)",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test<",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Name (Blacklist => >)",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test>",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Name (Blacklist => {)",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test{",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Name (Blacklist => })",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test}",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Emails (Required)",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test",
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"emails": "cannot be blank",
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Emails (isEmail => 0)",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test",
				Emails: []string{
					"unit_test0",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"emails": map[string]any{
							"0": "must be a valid email address",
						},
					},
				},
			},
		},
		{
			"Create User => Failed Validation => Emails (isEmail => 1)",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"emails": map[string]any{
							"1": "must be a valid email address",
						},
					},
				},
			},
		},
		{
			"Create User => Duplicate Email",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test0@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 500,
				BodyPart: Response{
					Code:        "0500",
					Description: "INTERNAL_SERVER_ERROR",
				},
			},
		},
		{
			"Create User => Success",
			Request{
				Method: http.MethodPost,
				Url:    "/api/v1/user",
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test",
				Emails: []string{
					"unit_test0@email.com",
					"unit_test1@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 201,
				BodyPart: Response{
					Code:        "0201",
					Description: "CREATED",
					Data: map[string]any{
						"deletedAt": nil,
						"name":      "Unit Test",
						"emails": []map[string]any{
							{
								"deletedAt": nil,
								"email":     "unit_test0@email.com",
							},
							{
								"deletedAt": nil,
								"email":     "unit_test1@email.com",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.TestName, func(t *testing.T) {
			c, recorder := PrepareContextFromTestCase(test)

			if assert.NoError(t, test.HandlerFunc(c)) {
				assert.Equal(t, test.Expected.StatusCode, recorder.Code)

				var recorderResponse Response
				json.Unmarshal(recorder.Body.Bytes(), &recorderResponse)

				assert.Equal(t, test.Expected.BodyPart.Code, recorderResponse.Code)

				assert.Equal(t, test.Expected.BodyPart.Description, recorderResponse.Description)

				if test.Expected.StatusCode != 201 {
					assert.Equal(t, test.Expected.BodyPart.Data, recorderResponse.Data)
				} else {
					UserDataTest(t, test.Expected.BodyPart.Data, recorderResponse.Data)
				}
			}
		})
	}
}

func TestReadUser(t *testing.T) {
	var handlerFunc = func(c echo.Context) error {
		return userHandlerFunc.Read(c)
	}

	cases := []TestCase{
		{
			"Read User => Failed Validation => Page (isDigit)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?page=A",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"page": "must contain digits only",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Limit (isDigit)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?limit=A",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"limit": "must contain digits only",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Order By (In)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?orderBy=A",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"orderBy": "must be a valid value",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Sort By (In)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?sortBy=A",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"sortBy": "must be a valid value",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Search (Blacklist => ')",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?search=A'",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"search": "the search contains unsafe characters",
					},
				},
			},
		},
		{
			`Read User => Failed Validation => Search (Blacklist => ")`,
			Request{
				Method: http.MethodGet,
				Url:    `/api/v1/user?search=A"`,
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"search": "the search contains unsafe characters",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Search (Blacklist => [)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?search=A[",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"search": "the search contains unsafe characters",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Search (Blacklist => ])",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?search=A]",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"search": "the search contains unsafe characters",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Search (Blacklist => <)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?search=A<",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"search": "the search contains unsafe characters",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Search (Blacklist => >)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?search=A>",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"search": "the search contains unsafe characters",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Search (Blacklist => {)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?search=A{",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"search": "the search contains unsafe characters",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Search (Blacklist => })",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?search=A}",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"search": "the search contains unsafe characters",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => Disable Calculate Total (In)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?disableCalculateTotal=A",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"disableCalculateTotal": "must be a valid value",
					},
				},
			},
		},
		{
			"Read User => Failed Validation => ID (isUUID)",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?id=A",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"id": "must be a valid UUID",
					},
				},
			},
		},
		{
			"Read User => Success => Without Query Param",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 200,
				BodyPart: Response{
					Code:        "0200",
					Description: "SUCCESS",
					Data:        map[string]any{},
				},
			},
		},
		{
			"Read User => Success",
			Request{
				Method: http.MethodGet,
				Url:    "/api/v1/user?page=1&limit=10&orderBy=name&sortBy=asc&search=Unit&disableCalculateTotal=true",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 200,
				BodyPart: Response{
					Code:        "0200",
					Description: "SUCCESS",
					Data:        map[string]any{},
				},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.TestName, func(t *testing.T) {
			c, recorder := PrepareContextFromTestCase(test)

			if assert.NoError(t, test.HandlerFunc(c)) {
				assert.Equal(t, test.Expected.StatusCode, recorder.Code)

				var recorderResponse Response
				json.Unmarshal(recorder.Body.Bytes(), &recorderResponse)

				assert.Equal(t, test.Expected.BodyPart.Code, recorderResponse.Code)

				assert.Equal(t, test.Expected.BodyPart.Description, recorderResponse.Description)

				if test.Expected.StatusCode != 200 {
					assert.Equal(t, test.Expected.BodyPart.Data, recorderResponse.Data)
				} else {
					recorderResponseDataBytes, err := json.Marshal(recorderResponse.Data)

					assert.Condition(t, func() bool {
						return err == nil
					}, "Failed to JSON Marshal for Recorder Response Data. Actual: %v", recorderResponse.Data)

					var recorderResponseData map[string]any

					json.Unmarshal(recorderResponseDataBytes, &recorderResponseData)

					assert.Condition(t, func() bool {
						return len(recorderResponseData["data"].([]any)) != 0
					}, "Expected the Data more than Zero. Actual: %v", len(recorderResponseData["data"].([]any)))

					dataSliceBytes, err := json.Marshal(recorderResponseData["data"])

					assert.Condition(t, func() bool {
						return err == nil
					}, "Failed to JSON Marshal for Recorder Response Data => Data. Actual: %v", recorderResponseData["data"])

					var dataSlice []map[string]any

					json.Unmarshal(dataSliceBytes, &dataSlice)

					for _, v := range dataSlice {
						UserDataTest(t, test.Expected.BodyPart.Data, v)
					}

					assert.Condition(t, func() bool {
						_, ok := recorderResponseData["total"].(float64)

						return ok
					}, "Expected the Total in Float64 form. Actual: %T", recorderResponseData["total"])

					assert.Condition(t, func() bool {
						_, ok := recorderResponseData["nextPage"].(bool)

						return ok
					}, "Expected the Next Page in Boolean form. Actual: %T", recorderResponseData["nextPage"])
				}
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	headers := []Header{
		{
			Key:   "Content-Type",
			Value: "application/json",
		},
	}

	var handlerFunc = func(c echo.Context) error {
		return userHandlerFunc.Update(c)
	}

	cases := []TestCase{
		{
			"Update User => Failed Validation => ID (Required)",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user",
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"id": "cannot be blank",
					},
				},
			},
		},
		{
			"Update User => Failed Validation => ID (isUUID)",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/A",
				PathParam: &PathParam{
					Name:  "id",
					Value: "A",
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"id": "must be a valid UUID",
					},
				},
			},
		},
		{
			"Update User => Failed Validation => Name (Blacklist => ')",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update'",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			`Update User => Failed Validation => Name (Blacklist => ")`,
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: `Unit Test Update"`,
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Update User => Failed Validation => Name (Blacklist => [)",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update[",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Update User => Failed Validation => Name (Blacklist => ])",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update]",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Update User => Failed Validation => Name (Blacklist => <)",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update<",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Update User => Failed Validation => Name (Blacklist => >)",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update>",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Update User => Failed Validation => Name (Blacklist => {)",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update{",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Update User => Failed Validation => Name (Blacklist => })",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update}",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"name": "the name contains unsafe characters",
					},
				},
			},
		},
		{
			"Update User => Failed Validation => Emails (isEmail => 0)",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update",
				Emails: []string{
					"unit_test0_update",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"emails": map[string]any{
							"0": "must be a valid email address",
						},
					},
				},
			},
		},
		{
			"Update User => Failed Validation => Emails (isEmail => 1)",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"emails": map[string]any{
							"1": "must be a valid email address",
						},
					},
				},
			},
		},
		{
			"Update User => Duplicate Email",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.UpdateUserRequest{
				Name: "Unit Test Update",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test0_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 500,
				BodyPart: Response{
					Code:        "0500",
					Description: "INTERNAL_SERVER_ERROR",
				},
			},
		},
		{
			"Update User => Success",
			Request{
				Method: http.MethodPatch,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			&headers,
			types.CreateUserRequest{
				Name: "Unit Test Update",
				Emails: []string{
					"unit_test0_update@email.com",
					"unit_test1_update@email.com",
				},
			},
			handlerFunc,
			ExpectedResponse{
				StatusCode: 200,
				BodyPart: Response{
					Code:        "0200",
					Description: "SUCCESS",
				},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.TestName, func(t *testing.T) {
			c, recorder := PrepareContextFromTestCase(test)

			if assert.NoError(t, test.HandlerFunc(c)) {
				assert.Equal(t, test.Expected.StatusCode, recorder.Code)

				var recorderResponse Response
				json.Unmarshal(recorder.Body.Bytes(), &recorderResponse)

				assert.Equal(t, test.Expected.BodyPart.Code, recorderResponse.Code)

				assert.Equal(t, test.Expected.BodyPart.Description, recorderResponse.Description)

				assert.Equal(t, test.Expected.BodyPart.Data, recorderResponse.Data)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	var handlerFunc = func(c echo.Context) error {
		return userHandlerFunc.Delete(c)
	}

	cases := []TestCase{
		{
			"Delete User => Failed Validation => ID (Required)",
			Request{
				Method: http.MethodDelete,
				Url:    "/api/v1/user",
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"id": "cannot be blank",
					},
				},
			},
		},
		{
			"Delete User => Failed Validation => ID (isUUID)",
			Request{
				Method: http.MethodDelete,
				Url:    "/api/v1/user/A",
				PathParam: &PathParam{
					Name:  "id",
					Value: "A",
				},
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 400,
				BodyPart: Response{
					Code:        "0400",
					Description: "BAD_REQUEST",
					Data: map[string]any{
						"id": "must be a valid UUID",
					},
				},
			},
		},
		{
			"Delete User => Success",
			Request{
				Method: http.MethodDelete,
				Url:    "/api/v1/user/" + id,
				PathParam: &PathParam{
					Name:  "id",
					Value: id,
				},
			},
			nil,
			nil,
			handlerFunc,
			ExpectedResponse{
				StatusCode: 200,
				BodyPart: Response{
					Code:        "0200",
					Description: "SUCCESS",
				},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.TestName, func(t *testing.T) {
			c, recorder := PrepareContextFromTestCase(test)

			if assert.NoError(t, test.HandlerFunc(c)) {
				assert.Equal(t, test.Expected.StatusCode, recorder.Code)

				var recorderResponse Response
				json.Unmarshal(recorder.Body.Bytes(), &recorderResponse)

				assert.Equal(t, test.Expected.BodyPart.Code, recorderResponse.Code)

				assert.Equal(t, test.Expected.BodyPart.Description, recorderResponse.Description)

				assert.Equal(t, test.Expected.BodyPart.Data, recorderResponse.Data)
			}
		})
	}
}
