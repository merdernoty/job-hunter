package app

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	httpResponse "github.com/merdernoty/job-hunter/pkg/http"
)

type Router struct {
	server *Server
}

func NewRouter(server *Server) *Router {
	return &Router{server: server}
}

func (r *Router) RegisterRoutes() {
	api := r.server.Echo().Group("/api")

	api.GET("/health", r.healthCheck)
	v1 := api.Group("/v1")
	r.registerV1Routes(v1)
}

func (r *Router) healthCheck(c echo.Context) error {
	healthData := map[string]interface{}{
		"status":    "healthy",
		"version":   r.server.config.Server.AppVersion,
		"mode":      r.server.config.Server.Mode,
		"timestamp": time.Now(),
	}

	return httpResponse.SuccessResponse(c, healthData, "Service is healthy")
}

func (r *Router) registerV1Routes(g *echo.Group) {
	g.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Test endpoint",
		})
	})
}
