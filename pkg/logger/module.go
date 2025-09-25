package logger

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewLogger),
	fx.Invoke(func(l Logger) {
		l.Info("Logger initialized successfully")
	}),
)