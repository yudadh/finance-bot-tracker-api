package pagination

import "math"

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type Params struct {
	Page    int
	PerPage int
}

func (p Params) Normalize() Params {
	if p.Page < 1 {
		p.Page = 1
	}

	if p.PerPage < 10 {
		p.PerPage = 10
	}

	return p
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p Params) Limit() int {
	return p.PerPage
}

func NewMeta(param Params, total int64) Meta {
	totalPages := int(math.Ceil(float64(total) / float64(param.PerPage)))
	return Meta{
		Page:    param.Page,
		PerPage: param.PerPage,
		Total:   total,
		TotalPages: totalPages,
	}
}