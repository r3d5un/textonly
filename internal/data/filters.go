package data

type Metadata struct {
	CurrentPage  int    `json:"current_page,omitempty"`
	PageSize     int    `json:"page_size,omitempty"`
	FirstPage    int    `json:"first_page,omitempty"`
	LastPage     int    `json:"last_page,omitempty"`
	TotalRecords int    `json:"total_records,omitempty"`
	OrderBy      string `json:"order_by,omitempty"`
}
