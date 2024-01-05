package data

import (
	"math"
	"strings"
	"time"

	"textonly.islandwind.me/internal/validator"
)

type Metadata struct {
	CurrentPage  int    `json:"current_page,omitempty"`
	PageSize     int    `json:"page_size,omitempty"`
	FirstPage    int    `json:"first_page,omitempty"`
	LastPage     int    `json:"last_page,omitempty"`
	TotalRecords int    `json:"total_records,omitempty"`
	OrderBy      string `json:"order_by,omitempty"`
}

type Filters struct {
	Page            int        `json:"page,omitempty"`
	PageSize        int        `json:"page_size,omitempty"`
	ID              *int       `json:"id,omitempty"`
	UserID          int        `json:"user_id,omitempty"`
	Title           string     `json:"title,omitempty"`
	Lead            string     `json:"lead,omitempty"`
	Post            string     `json:"post,omitempty"`
	CreatedFrom     *time.Time `json:"created_from,omitempty"`
	CreatedTo       *time.Time `json:"created_to,omitempty"`
	LastUpdatedFrom *time.Time `json:"last_updated_from,omitempty"`
	LastUpdatedTo   *time.Time `json:"last_updated_to,omitempty"`
	Name            string     `json:"name,omitempty"`
	SocialPlatform  string     `json:"social_platform,omitempty"`
	OrderBy         []string   `json:"order_by"`
	OrderBySafeList []string   `json:"order_by_safe_list"`
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 50_000, "page_size", "must be a maximum of 50,000")

	orderByParam, isPermitted := validator.PermittedValues(f.OrderBy, f.OrderBySafeList)
	v.Check(isPermitted, orderByParam, "invalid order_by parameter")
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func calculateMetadata(totalRecords, page, pageSize int, orderBySlice []string) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
		OrderBy:      strings.Join(orderBySlice, ","),
	}
}
