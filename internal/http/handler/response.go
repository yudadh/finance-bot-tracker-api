package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HandlerFunc func(c *gin.Context) error

type Response[T any] struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type ResponseWithMeta[T any, M any] struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    M      `json:"meta,omitempty"`
}

func Handle(logger *slog.Logger, fn HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := fn(ctx); err != nil {
			HandleError(ctx, err)

			if ctx.Writer.Status() >= http.StatusInternalServerError {
				logger.ErrorContext(ctx.Request.Context(),
					"http request failed",
					"method", ctx.Request.Method,
					"path", ctx.Request.URL.Path,
					"status", ctx.Writer.Status(),
					"err", err,
				)
			}
		}
	}
}

func ResponseSuccess[T any](
	c *gin.Context,
	statusCode int,
	message string,
	data T,
) {
	c.JSON(statusCode, Response[T]{
		Error:   false,
		Message: message,
		Data:    data,
	})
}

func ResponseSuccessWithMeta[T any, M any](
	c *gin.Context,
	statusCode int,
	message string,
	data T,
	meta M,
) {
	c.JSON(statusCode, ResponseWithMeta[T, M]{
		Error:   false,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func ResponseError(c *gin.Context, statusCode int, message string, fieldErrors []ValidationFieldError) {
	c.JSON(statusCode, ErrorResponse{
		Error:   true,
		Message: message,
		Errors:  fieldErrors,
	})
}
