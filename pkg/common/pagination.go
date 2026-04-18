package common

// Pagination holds page/limit from query params and total from repo.
// Pass as *Pagination to services/repos — repo sets Total after counting.
type Pagination struct {
	Page  int   `json:"page"  query:"page"`
	Limit int   `json:"limit" query:"limit"`
	Total int64 `json:"total"`
}

// Offset returns the SQL offset for the current page
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}
