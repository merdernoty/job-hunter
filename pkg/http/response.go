package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func SuccessResponse(c echo.Context, data interface{}, message ...string) error {
	msg := "Success"
	if len(message) > 0 {
		msg = message[0]
	}

	return c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   msg,
		Data:      data,
		Timestamp: time.Now(),
	})
}

func CreatedResponse(c echo.Context, data interface{}, message ...string) error {
	msg := "Created successfully"
	if len(message) > 0 {
		msg = message[0]
	}

	return c.JSON(http.StatusCreated, Response{
		Success:   true,
		Message:   msg,
		Data:      data,
		Timestamp: time.Now(),
	})
}

func ErrorResponse(c echo.Context, statusCode int, code, message string, details ...string) error {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}

	return c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: detail,
		},
		Timestamp: time.Now(),
	})
}

func BadRequestResponse(c echo.Context, message string, details ...string) error {
	return ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", message, details...)
}

func NotFoundResponse(c echo.Context, message string, details ...string) error {
	return ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", message, details...)
}

func InternalServerErrorResponse(c echo.Context, message string, details ...string) error {
	return ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, details...)
}

func UnauthorizedResponse(c echo.Context, message string, details ...string) error {
	return ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message, details...)
}

func ForbiddenResponse(c echo.Context, message string, details ...string) error {
	return ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", message, details...)
}