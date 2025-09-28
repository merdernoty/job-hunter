package user

import (
	"github.com/merdernoty/job-hunter/internal/users/domain"
	"github.com/merdernoty/job-hunter/internal/users/repository"
	"github.com/merdernoty/job-hunter/internal/users/service"
	"go.uber.org/fx"
)

var Module = fx.Module("user",
	fx.Provide(
		fx.Annotate(
			repository.NewUserRepository,
			fx.As(new(domain.UserRepository)),
		),
		fx.Annotate(
			repository.NewUserDailyViewRepository,
			fx.As(new(domain.UserDailyViewRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			service.NewUserService,
			fx.As(new(domain.UserService)),
		),
		service.NewAvatarService,
	),
)
