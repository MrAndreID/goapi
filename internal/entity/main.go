package entity

import (
	"errors"
	"regexp"

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
	Records  any   `json:"records"`
	Total    int64 `json:"total"`
	NextPage bool  `json:"nextPage"`
}

type MainResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Data        any    `json:"data"`
}

func BlacklistValidation(field string) validation.RuleFunc {
	return func(value interface{}) error {
		val, ok := value.(string)

		if !ok {
			return errors.New("the " + field + " is not a string")
		}

		if val == "" {
			return nil
		}

		match, _ := regexp.MatchString(`^[^'"\[\]<>\{\}]+$`, val)

		if !match {
			return errors.New("the " + field + " contains unsafe characters")
		}

		return nil
	}
}
