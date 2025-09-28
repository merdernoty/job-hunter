package controller

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/merdernoty/job-hunter/internal/users/domain"
	httpResponse "github.com/merdernoty/job-hunter/pkg/http"
)

type UserController struct {
	userService domain.UserService
}

func NewUserController(userService domain.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (ctrl *UserController) RegisterRoutes(rg *echo.Group) {
	// Auth routes
	auth := rg.Group("/auth")
	auth.POST("/telegram", ctrl.authTelegram)

	// User routes
	users := rg.Group("/users")
	users.GET("/:id", ctrl.getByID)
	users.PUT("/:id", ctrl.update)

	// Profile routes
	users.GET("/me", ctrl.getProfile)
	users.PUT("/me", ctrl.updateProfile)
	users.PUT("/me/avatar", ctrl.updateAvatar)
	users.DELETE("/me/avatar", ctrl.deleteAvatar)

	// Match routes
	users.GET("/:viewerID/random", ctrl.getRandomUser)
}

func (ctrl *UserController) authTelegram(c echo.Context) error {
	var req domain.TelegramAuthRequest

	if err := httpResponse.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, token, err := ctrl.userService.AuthFromTelegram(req.InitData)
	if err != nil {
		switch err.Error() {
		case "invalid telegram data":
			return httpResponse.BadRequestResponse(c, "Invalid Telegram data")
		case "database error":
			return httpResponse.InternalServerErrorResponse(c, "Database error")
		default:
			return httpResponse.UnauthorizedResponse(c, "Authentication failed")
		}
	}

	return httpResponse.SuccessResponse(c, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func (ctrl *UserController) getByID(c echo.Context) error {
	idParam := c.Param("id")
	if idParam == "" {
		return httpResponse.BadRequestResponse(c, "User ID is required")
	}

	userID, err := uuid.Parse(idParam)
	if err != nil {
		return httpResponse.BadRequestResponse(c, "Invalid user ID format")
	}

	user, err := ctrl.userService.GetUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			return httpResponse.NotFoundResponse(c, "User not found")
		}
		return httpResponse.InternalServerErrorResponse(c, "Failed to retrieve user")
	}

	return httpResponse.SuccessResponse(c, user)
}

func (ctrl *UserController) update(c echo.Context) error {
	idParam := c.Param("id")
	if idParam == "" {
		return httpResponse.BadRequestResponse(c, "User ID is required")
	}

	userID, err := uuid.Parse(idParam)
	if err != nil {
		return httpResponse.BadRequestResponse(c, "Invalid user ID format")
	}

	var req domain.UpdateUserRequest
	if err := httpResponse.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := ctrl.userService.UpdateUser(userID, req)
	if err != nil {
		if err.Error() == "user not found" {
			return httpResponse.NotFoundResponse(c, "User not found")
		}
		if err.Error() == "no fields to update" {
			return httpResponse.BadRequestResponse(c, "No fields to update")
		}
		return httpResponse.InternalServerErrorResponse(c, "Failed to update user")
	}

	return httpResponse.SuccessResponse(c, user)
}

func (ctrl *UserController) getProfile(c echo.Context) error {
	// TODO: Получить ID пользователя из JWT токена
	// Пока получаем из заголовка для тестирования
	userIDStr := c.Request().Header.Get("X-User-ID")
	if userIDStr == "" {
		return httpResponse.UnauthorizedResponse(c, "Authentication required")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return httpResponse.BadRequestResponse(c, "Invalid user ID in header")
	}

	user, err := ctrl.userService.GetUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			return httpResponse.NotFoundResponse(c, "User not found")
		}
		return httpResponse.InternalServerErrorResponse(c, "Failed to retrieve profile")
	}

	return httpResponse.SuccessResponse(c, user)
}

func (ctrl *UserController) updateProfile(c echo.Context) error {
	userIDStr := c.Request().Header.Get("X-User-ID")
	if userIDStr == "" {
		return httpResponse.UnauthorizedResponse(c, "Authentication required")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return httpResponse.BadRequestResponse(c, "Invalid user ID in header")
	}

	var req domain.UpdateUserRequest
	if err := httpResponse.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := ctrl.userService.UpdateUser(userID, req)
	if err != nil {
		if err.Error() == "user not found" {
			return httpResponse.NotFoundResponse(c, "User not found")
		}
		if err.Error() == "no fields to update" {
			return httpResponse.BadRequestResponse(c, "No fields to update")
		}
		return httpResponse.InternalServerErrorResponse(c, "Failed to update profile")
	}

	return httpResponse.SuccessResponse(c, user)
}
func (ctrl *UserController) updateAvatar(c echo.Context) error {
	userIDStr := c.Request().Header.Get("X-User-ID")
	if userIDStr == "" {
		return httpResponse.UnauthorizedResponse(c, "Authentication required")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return httpResponse.BadRequestResponse(c, "Invalid user ID in header")
	}

	file, header, err := c.Request().FormFile("avatar")
	if err != nil {
		return httpResponse.BadRequestResponse(c, "No avatar file provided")
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		if strings.HasSuffix(strings.ToLower(header.Filename), ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(strings.ToLower(header.Filename), ".gif") {
			contentType = "image/gif"
		} else {
			contentType = "image/jpeg"
		}
	}

	avatarURL, err := ctrl.userService.UpdateUserAvatar(userID, file, header.Filename, header.Size, contentType)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "user not found"):
			return httpResponse.NotFoundResponse(c, "User not found")
		case strings.Contains(err.Error(), "invalid file type"):
			return httpResponse.BadRequestResponse(c, "Invalid file type: only images are allowed")
		case strings.Contains(err.Error(), "file too large"):
			return httpResponse.BadRequestResponse(c, "Avatar file too large: maximum size is 2MB")
		default:
			return httpResponse.InternalServerErrorResponse(c, "Failed to update avatar")
		}
	}

	return httpResponse.SuccessResponse(c, map[string]interface{}{
		"avatar_url": avatarURL,
	}, "Avatar updated successfully")
}

func (ctrl *UserController) deleteAvatar(c echo.Context) error {
	userIDStr := c.Request().Header.Get("X-User-ID")
	if userIDStr == "" {
		return httpResponse.UnauthorizedResponse(c, "Authentication required")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return httpResponse.BadRequestResponse(c, "Invalid user ID in header")
	}

	err = ctrl.userService.DeleteUserAvatar(userID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "user not found"):
			return httpResponse.NotFoundResponse(c, "User not found")
		case strings.Contains(err.Error(), "has no avatar"):
			return httpResponse.BadRequestResponse(c, "User has no avatar to delete")
		default:
			return httpResponse.InternalServerErrorResponse(c, "Failed to delete avatar")
		}
	}

	return httpResponse.SuccessResponse(c, nil)
}

func (ctrl *UserController) getRandomUser(c echo.Context) error {
	viewerIDStr := c.Param("viewerID")
	if viewerIDStr == "" {
		return httpResponse.BadRequestResponse(c, "Viewer ID is required")
	}

	viewerID, err := uuid.Parse(viewerIDStr)
	if err != nil {
		return httpResponse.BadRequestResponse(c, "Invalid viewer ID format")
	}

	randomUser, err := ctrl.userService.GetRandomUser(viewerID)
	if err != nil {
		if err.Error() == "no more users available today" {
			return httpResponse.SuccessResponse(c, nil, "На сегодня все пользователи просмотрены")
		}
		return httpResponse.InternalServerErrorResponse(c, "Failed to get random user")
	}

	return httpResponse.SuccessResponse(c, randomUser)
}
