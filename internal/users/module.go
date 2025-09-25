package user

import (
	"go.uber.org/fx"
	"github.com/merdernoty/job-hunter/internal/users/domain"
	"github.com/merdernoty/job-hunter/internal/users/repository"
	"github.com/merdernoty/job-hunter/internal/users/service"
)

var Module = fx.Module("user",
	fx.Provide(
		fx.Annotate(
			repository.NewUserRepository,
			fx.As(new(domain.UserRepository)),
		),
	),

	fx.Provide(
		fx.Annotate(
			service.NewUserService,
			fx.As(new(domain.UserService)),
		),
	),
)