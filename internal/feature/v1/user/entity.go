package user

import (
	"github.com/MrAndreID/goapi/v2/internal/entity"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type CreateData struct {
	Name   string   `json:"name"`
	Emails []string `json:"emails"`
}

type ReadData struct {
	entity.PaginatorRequest
	ID string `query:"id" json:"id"`
}

type UpdateData struct {
	ID     string   `param:"id" json:"id"`
	Name   string   `json:"name"`
	Emails []string `json:"emails"`
}

type DeleteData struct {
	ID string `param:"id" json:"id"`
}

func (r CreateData) Validate() interface{} {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.By(entity.BlacklistValidation("name"))),
		validation.Field(&r.Emails, validation.Required, validation.Each(is.Email)),
	)
}

func (r ReadData) Validate() interface{} {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Page, is.Digit),
		validation.Field(&r.Limit, is.Digit),
		validation.Field(&r.OrderBy, validation.In("id", "name", "createdAt", "updatedAt")),
		validation.Field(&r.SortBy, validation.In("asc", "desc")),
		validation.Field(&r.Search, validation.By(entity.BlacklistValidation("search"))),
		validation.Field(&r.DisableCalculateTotal, validation.In("true", "false")),
		validation.Field(&r.ID, is.UUID),
	)
}

func (r UpdateData) Validate() interface{} {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
		validation.Field(&r.Name, validation.By(entity.BlacklistValidation("name"))),
		validation.Field(&r.Emails, validation.Each(is.Email)),
	)
}

func (r DeleteData) Validate() interface{} {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	)
}
