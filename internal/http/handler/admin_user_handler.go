package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/pagination"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type UserAdminService interface {
	Register(ctx context.Context, adminUserInput service.RegisterAdminUserInput) (*service.RegisterAdminUserResult, error)
	Login(ctx context.Context, email string, password string) (string, error)
	FindByID(ctx context.Context, id uint64) (*service.FindByIDResult, error)
}

type UserService interface {
	FindAll(ctx context.Context, query service.FindAllUsersQuery) ([]service.FindAllUserResult, *pagination.Meta, error)
}

type UserAdminHandler struct {
	userAdminService UserAdminService
	userService      UserService
	logger           *slog.Logger
}

func NewUserAdminHandler(
	userAdminService UserAdminService,
	userService UserService,
	logger *slog.Logger,
) *UserAdminHandler {
	return &UserAdminHandler{
		userAdminService: userAdminService,
		userService:      userService,
		logger:           logger,
	}
}

func (h *UserAdminHandler) Register(c *gin.Context) error {
	var request RegisterAdminUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		return err
	}

	adminUser := service.RegisterAdminUserInput{
		Email:    request.Email,
		Password: request.Password,
		Name:     request.Name,
	}

	result, err := h.userAdminService.Register(c.Request.Context(), adminUser)
	if err != nil {
		return err
	}

	ResponseSuccess(c, http.StatusCreated, "user admin created successfully", RegisterAdminUserResponse{
		ID:     result.ID,
		Email:  result.Email,
		Name:   result.Name,
		Status: string(result.Status),
	})
	return nil
}

func (h *UserAdminHandler) Login(c *gin.Context) error {
	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		return err
	}

	token, err := h.userAdminService.Login(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		return err
	}

	ResponseSuccess(c, http.StatusOK, "login success", LoginResponse{Token: token})
	return nil
}

func (h *UserAdminHandler) GetMe(c *gin.Context) error {
	value, ok := c.Get("admin_user_id")
	if !ok {
		return errors.New("missing authentication context")
	}

	adminUserID, ok := value.(uint64)
	if !ok {
		return errors.New("invalid authentication context")
	}

	result, err := h.userAdminService.FindByID(c.Request.Context(), adminUserID)
	if err != nil {
		return err
	}

	ResponseSuccess(c, http.StatusOK, "user admin successfully fetched", GetMeResponse{
		ID:     result.ID,
		Email:  result.Email,
		Name:   result.Name,
		Status: string(result.Status),
	})
	return nil
}

func (h *UserAdminHandler) GetAllUsers(c *gin.Context) error {
	var query ListUsersQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}

	params := pagination.Params{
		Page:    query.Page,
		PerPage: query.PerPage,
	}.Normalize()

	results, paginationMeta, err := h.userService.FindAll(c.Request.Context(), service.FindAllUsersQuery{
		Param:  params,
		Status: query.Status,
		Search: query.Search,
	})

	if err != nil {
		return err
	}

	data := []ListUsersResponse{}

	for _, res := range results {
		data = append(data, ListUsersResponse{
			ID:               res.ID,
			TelegramID:       res.TelegramID,
			TelegramUsername: res.TelegramUsername,
			FirstName:        res.FirstName,
			LastName:         res.LastName,
			Timezone:         res.Timezone,
			LanguageCode:     res.LanguageCode,
			Status:           string(res.Status),
		})
	}

	ResponseSuccessWithMeta(c, http.StatusOK, "users successfully fetched", data, paginationMeta)
	return nil
}
