package app

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

type Router struct {
	server *Server
}

func NewRouter(server *Server) *Router {
	return &Router{server: server}
}

func (r *Router) RegisterRoutes() {
	api := r.server.Echo().Group("/api")
	
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"version": r.server.config.Server.AppVersion,
		})
	})
		v1 := api.Group("/v1")
	r.registerV1Routes(v1)
}

func (r *Router) registerV1Routes(g *echo.Group) {
	g.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Test endpoint",
		})
	})
}