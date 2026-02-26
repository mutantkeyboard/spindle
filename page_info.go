package spindle

import "fmt"

// SortOrder represents sort order.
type SortOrder string

const (
	ASC  SortOrder = "asc"
	DESC SortOrder = "desc"
)

// SortField represents a sort field with direction.
type SortField struct {
	Field string
	Order SortOrder
}

// SortOrderFromString returns a SortOrder from a string.
func SortOrderFromString(s string) SortOrder {
	switch s {
	case "asc":
		return ASC
	case "desc":
		return DESC
	default:
		return ASC
	}
}

// PageInfo contains pagination information.
type PageInfo struct {
	Page   int         `json:"page"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Sort   []SortField `json:"sort"`
}

// NewPageInfo creates a new PageInfo.
func NewPageInfo(page, limit, offset int, sort []SortField) *PageInfo {
	return &PageInfo{
		Page:   page,
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}
}

// Start returns the start index based on page/limit or offset.
func (p *PageInfo) Start() int {
	if p.Offset > 0 {
		return p.Offset
	}
	return (p.Page - 1) * p.Limit
}

// SortBy adds a sort field. Chainable.
func (p *PageInfo) SortBy(field string, order SortOrder) *PageInfo {
	p.Sort = append(p.Sort, SortField{Field: field, Order: order})
	return p
}

// NextPageURL returns the URL for the next page.
func (p *PageInfo) NextPageURL(baseURL string) string {
	return fmt.Sprintf("%s?page=%d&limit=%d", baseURL, p.Page+1, p.Limit)
}

// PreviousPageURL returns the URL for the previous page.
// Returns empty string if on page 1.
func (p *PageInfo) PreviousPageURL(baseURL string) string {
	if p.Page > 1 {
		return fmt.Sprintf("%s?page=%d&limit=%d", baseURL, p.Page-1, p.Limit)
	}
	return ""
}
