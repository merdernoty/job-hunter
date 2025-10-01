package controller

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/merdernoty/job-hunter/internal/users/domain"
	"github.com/merdernoty/job-hunter/internal/users/middleware"
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

func (ctrl *UserController) RegisterRoutes(rg *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	// Auth routes
	auth := rg.Group("/auth")
	auth.POST("/telegram", ctrl.authTelegram)

	// User routes
	users := rg.Group("/users", jwtMiddleware)
	users.GET("", ctrl.getUsers)
	users.GET("/:id", ctrl.getByID)
	users.PUT("/:id", ctrl.update)

	// Profile routes
	users.GET("/me", ctrl.getProfile)
	users.PUT("/me", ctrl.updateProfile)
	users.PUT("/me/avatar", ctrl.updateAvatar)
	users.DELETE("/me/avatar", ctrl.deleteAvatar)

	// Match routes
	users.GET("/random", ctrl.getRandomUser)
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
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return httpResponse.UnauthorizedResponse(c, "Authentication required")
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
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return httpResponse.UnauthorizedResponse(c, "Authentication required")
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
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return httpResponse.UnauthorizedResponse(c, "Authentication required")
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
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return httpResponse.UnauthorizedResponse(c, "Authentication required")
	}

	err := ctrl.userService.DeleteUserAvatar(userID)
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
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return httpResponse.UnauthorizedResponse(c, "Authentication required")
	}

	randomUser, err := ctrl.userService.GetRandomUser(userID)
	if err != nil {
		if err.Error() == "no more users available today" {
			return httpResponse.SuccessResponse(c, nil, "На сегодня все пользователи просмотрены")
		}
		return httpResponse.InternalServerErrorResponse(c, "Failed to get random user")
	}

	return httpResponse.SuccessResponse(c, randomUser)
}

func (ctrl *UserController) getUsers(c echo.Context) error {
	users, err := ctrl.userService.GetAllUsers()
	if err != nil {
		return httpResponse.InternalServerErrorResponse(c, "Failed to get users")
	}
	return httpResponse.SuccessResponse(c, users)
}
