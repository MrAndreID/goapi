package entity

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type PaginatorRequest struct {
	Page                  string `query:"page" json:"page"`
	Limit                 string `query:"limit" json:"limit"`
	OrderBy               string `query:"orderBy" json:"orderBy"`
	SortBy                string `query:"sortBy" json:"sortBy"`
	Search                string `query:"search" json:"search"`
	DisableCalculateTotal string `query:"disableCalculateTotal" json:"disableCalculateTotal"`
}

type PaginatorResponse struct {
	Records any    `json:"-"`
	Total   *int64 `json:"total"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
}

type MainResponse struct {
	Code    int    `json:"-"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    any    `json:"meta"`
	Error   any    `json:"error"`
}

func (mr MainResponse) JSON() (int, MainResponse) {
	if mr.Code == 0 {
		mr.Code = http.StatusNotFound

		mr.Message = "STATUS_CODE_NOT_FOUND"

		mr.Data = nil

		mr.Meta = nil

		mr.Error = []string{
			"status code not found",
		}
	}

	if !(mr.Code >= 400 && mr.Code < 600) {
		mr.Status = true
	}

	if mr.Message == "" {
		mr.Message = strings.ToUpper(strings.ReplaceAll(http.StatusText(mr.Code), " ", "_"))
	}

	return mr.Code, mr
}

func BlacklistValidation(field string) validation.RuleFunc {
	return func(value interface{}) error {
		val, ok := value.(string)

		if !ok {
			return errors.New("must be a valid string")
		}

		if val == "" {
			return nil
		}

		match, _ := regexp.MatchString(`^[^'"\[\]<>\{\}]+$`, val)

		if !match {
			return errors.New("must contains safe characters")
		}

		return nil
	}
}
