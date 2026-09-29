package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"github.com/yudadh/finance-bot-tracker-api/internal/pagination"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type CategoryService interface {
	Create(ctx context.Context, categoryInput service.CategoryInput) (*service.CategoryResult, error)
	FindByID(ctx context.Context, id uint64) (*service.CategoryResult, error)
	FindAll(ctx context.Context, query service.CategoryQuery) ([]service.CategoryResult, int64, error)
	Update(ctx context.Context, data service.UpdateCategoryInput) (*service.CategoryResult, error)
	HardDelete(ctx context.Context, id uint64) error
}

type CategoryHandler struct {
	categoryService CategoryService
}

func NewCategoryHandler(categoryService CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) Create(c *gin.Context) error {
	var body CategoryRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}

	result, err := h.categoryService.Create(c.Request.Context(), service.CategoryInput{
		Name: body.Name,
		Type: domain.TransactionType(body.Type),
		Keywords: body.Keywords,
		IsDefault: *body.IsDefault,
	})

	if err != nil {
		return err
	}

	ResponseSuccess(c, http.StatusCreated, "category successfully created", CategoryResponse{
		ID: result.ID,
		Name: result.Name,
		Type: string(result.Type),
		Keywords: result.Keywords,
		IsDefault: result.IsDefault,
	})

	return nil
}

func (h *CategoryHandler) FindByID(c *gin.Context) error {
	id, err := ParsePositiveID(c.Param("id"))
	if err != nil {
		return err
	}

	result, err := h.categoryService.FindByID(c.Request.Context(), id)

	ResponseSuccess(c, http.StatusOK, "category successfully fetched", &CategoryResponse{
		ID: result.ID,
		Name: result.Name,
		Type: string(result.Type),
		Keywords: result.Keywords,
		IsDefault: result.IsDefault,
	})

	return nil
}

func (h *CategoryHandler) FindAll(c *gin.Context) error {
	var query FindAllCategoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}
	
	result, total, err := h.categoryService.FindAll(c.Request.Context(), service.CategoryQuery{
		Param: pagination.Params{
			Page: query.Page,
			PerPage: query.PerPage,
		},
		TransactionType: domain.TransactionType(query.Type),
		Name: query.Name,
	})

	if err != nil {
		return err
	}

	data := make([]CategoryResponse, 0, len(result))

	for _, res := range result {
		data = append(data, CategoryResponse{
			ID: res.ID,
			Name: res.Name,
			Type: string(res.Type),
			Keywords: res.Keywords,
			IsDefault: res.IsDefault,
		})
	}

	ResponseSuccessWithMeta(c, http.StatusOK, "categories successfully fetched", data, pagination.NewMeta(
		pagination.Params{
			Page: query.Page,
			PerPage: query.PerPage,
		}, 
		total,
	))
	
	return nil
}

func (h *CategoryHandler) Update(c *gin.Context) error {
	id, err := ParsePositiveID(c.Param("id"))
	if err != nil {
		return err
	}

	var body CategoryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}

	result, err := h.categoryService.Update(c.Request.Context(), service.UpdateCategoryInput{
		ID: id,
		Name: body.Name,
		Type: domain.TransactionType(body.Type),
		Keywords: body.Keywords,
		IsDefault: *body.IsDefault,
	})

	if err != nil {
		return err
	}

	ResponseSuccess(c, http.StatusOK, "category successfully updated", CategoryResponse{
		ID: result.ID,
		Name: result.Name,
		Type: string(result.Type),
		Keywords: result.Keywords,
		IsDefault: result.IsDefault,
	})

	return nil
}

func (h *CategoryHandler) HardDelete(c *gin.Context) error {
	id, err := ParsePositiveID(c.Param("id"))
	if err != nil {
		return err
	}

	err = h.categoryService.HardDelete(c.Request.Context(), id)
	if err != nil {
		return err
	}

	ResponseSuccessWithoutData(c, http.StatusOK, "category successfully deleted permanently")
	return nil
}