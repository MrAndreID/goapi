package types

type PaginatorResponse struct {
	Data     any   `json:"data"`
	Total    int64 `json:"total"`
	NextPage bool  `json:"nextPage"`
}

type MainResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Data        any    `json:"data"`
}
