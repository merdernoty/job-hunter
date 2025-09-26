package config

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func NewViperConfig() (*viper.Viper, error) {
	return LoadConfig()
}

var Module = fx.Options(
	fx.Provide(NewViperConfig),
	fx.Provide(ParseConfig),
)
