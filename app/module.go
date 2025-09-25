package app

import (
	"github.com/merdernoty/job-hunter/internal/users/controller"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewServer),
	fx.Provide(controller.NewUserController),
	fx.Invoke(RegisterRoutes),
)