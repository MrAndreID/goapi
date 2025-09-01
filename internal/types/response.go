package types

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
