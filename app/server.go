package app

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/merdernoty/job-hunter/config"
	"github.com/merdernoty/job-hunter/pkg/logger"
	"go.uber.org/fx"
)

type Server struct {
	engine *echo.Echo
	config *config.Config
	logger logger.Logger
}

func NewServer(lc fx.Lifecycle, cfg *config.Config, log logger.Logger) *Server {
	engine := echo.New()
	
	engine.HideBanner = true
	engine.HidePort = true
	
	engine.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogMethod: true,
		LogLatency: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			log.Infof("REQUEST: %s %s - Status: %d - Latency: %v", 
				values.Method, values.URI, values.Status, values.Latency)
			return nil
		},
	}))
	
	engine.Use(middleware.Recover())

	s := &Server{
		engine: engine,
		config: cfg,
		logger: log,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				port := cfg.Server.Port
				if port == "" {
					port = ":8080"
				}
				log.Infof("Starting REST server on %s", port)
				log.Infof("Server mode: %s", cfg.Server.Mode)
				log.Infof("Server version: %s", cfg.Server.AppVersion)
				
				if err := engine.Start(port); err != nil {
					log.Fatalf("Server failed to start: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping server gracefully...")
			return s.engine.Shutdown(ctx)
		},
	})

	return s
}

func (s *Server) Echo() *echo.Echo {
	return s.engine
}

func (s *Server) Config() *config.Config {
	return s.config
}

func (s *Server) Logger() logger.Logger {
	return s.logger
}
