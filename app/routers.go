package app

import (
	"github.com/labstack/echo/v4"
	"github.com/merdernoty/job-hunter/internal/users/controller"
	httpResponse "github.com/merdernoty/job-hunter/pkg/http"
)

func RegisterRoutes(
	s *Server,
	userCtrl *controller.UserController,
) {
	s.Echo().GET("/api/health", healthCheck(s))
	// API v1
	api := s.Echo().Group("/api/v1")
	userCtrl.RegisterRoutes(api)
}

func healthCheck(s *Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		return httpResponse.SuccessResponse(c, map[string]interface{}{
			"status":  "healthy",
			"service": "job-hunter-api",
			"version": s.config.Server.AppVersion,
		}, "Service is running")
	}
}
