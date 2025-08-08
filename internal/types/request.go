package types

type CreateUserRequest struct {
	Name   string   `json:"name"`
	Emails []string `json:"emails"`
}

type PaginatorRequest struct {
	Page                  string `query:"page" json:"page"`
	Limit                 string `query:"limit" json:"limit"`
	OrderBy               string `query:"orderBy" json:"orderBy"`
	SortBy                string `query:"sortBy" json:"sortBy"`
	Search                string `query:"search" json:"search"`
	DisableCalculateTotal string `query:"disableCalculateTotal" json:"disableCalculateTotal"`
}

type ReadUserRequest struct {
	PaginatorRequest
	ID string `query:"id" json:"id"`
}

type UpdateUserRequest struct {
	ID     string   `param:"id" json:"id"`
	Name   string   `json:"name"`
	Emails []string `json:"emails"`
}

type DeleteUserRequest struct {
	ID string `param:"id" json:"id"`
}
