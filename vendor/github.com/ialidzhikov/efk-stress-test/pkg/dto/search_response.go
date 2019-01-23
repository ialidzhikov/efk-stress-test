package dto

type SearchResponse struct {
	Hits Hits `json:"hits"`
}

type Hits struct {
	Total uint64 `json:"total"`
}
