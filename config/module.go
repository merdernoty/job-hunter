package config

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func NewViperConfig() (*viper.Viper, error) {
	configNames := []string{"config", "app", "application"}
	
	for _, name := range configNames {
		if v, err := LoadConfig(name); err == nil {
			return v, nil
		}
	}
	return LoadConfig("config")
}

var Module = fx.Options(
	fx.Provide(NewViperConfig),
	fx.Provide(ParseConfig),
)