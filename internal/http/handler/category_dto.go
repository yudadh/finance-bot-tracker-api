package handler

type CategoryRequest struct {
	Name string `json:"name" binding:"required,min=1"`
	Type string `json:"type" binding:"required,oneof=expense income"`
	Keywords []string `json:"keywords" binding:"required,gt=0"`
	IsDefault *bool `json:"is_default" binding:"required"`
}

type CategoryResponse struct {
	ID uint64 `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Keywords []string `json:"keywords"`
	IsDefault bool `json:"is_default"`
}

type FindAllCategoryQuery struct {
	Page int `form:"page,default=1" binding:"gte=1"`
	PerPage int `form:"per_page,default=10" binding:"gte=10"`
	Type string `form:"type" binding:"omitempty,oneof=expense income"`
	Name string `form:"name" binding:"omitempty"`
}