package handlers

type Pagination struct {
	RecordsPerPage int `json:"records_per_page"`
	PageIndex      int `json:"page_index"`
}
