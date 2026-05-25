package http

type stopListRequest struct {
	Query string `json:"query"`
}

type stopListResponse struct {
	Items []string `json:"items"`
}
