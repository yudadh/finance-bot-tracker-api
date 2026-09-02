package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yudadh/finance-bot-tracker-api/internal/service"
)

type UserAdminService interface {
	Register(ctx context.Context, adminUserInput service.RegisterAdminUserInput) (*service.RegisterAdminUserResult, error)
	Login(ctx context.Context, email string, password string) (string, error)
	FindByID(ctx context.Context, id uint64) (*service.FindByIDResult, error)
}

type UserAdminHandler struct {
	userAdminService UserAdminService
	logger *slog.Logger
}

func NewUserAdminHandler(userAdminService UserAdminService, logger *slog.Logger) *UserAdminHandler {
	return &UserAdminHandler{
		userAdminService: userAdminService,
		logger: logger,
	}
}

func (h *UserAdminHandler) Register(c *gin.Context) error {
	var request RegisterAdminUserRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		return err
	}
	
	adminUser := service.RegisterAdminUserInput{
		Email: request.Email,
		Password: request.Password,
		Name: request.Name,
	}

	result, err := h.userAdminService.Register(c.Request.Context(), adminUser)
	if err != nil {
		return err
	}

	ResponseSuccess(c, http.StatusCreated, "user admin created successfully", RegisterAdminUserResponse{
		ID: result.ID,
		Email: result.Email,
		Name: result.Name,
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
		ID: result.ID,
		Email: result.Email,
		Name: result.Name,
		Status: string(result.Status),
	})
	return nil
}