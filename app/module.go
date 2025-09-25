package app

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewServer),
	fx.Provide(NewRouter),
	fx.Invoke(func(router *Router) {
		router.RegisterRoutes()
	}),
)