package bot

import (
	"context"

	"go.uber.org/fx"
	"github.com/merdernoty/job-hunter/config"
	"github.com/merdernoty/job-hunter/pkg/logger"
)

func NewBotFromConfig(cfg *config.Config, logger logger.Logger) (*Bot, error) {
	return NewBot(cfg.Bot.Token, cfg.Bot.WebAppURL, logger)
}

var Module = fx.Module("bot",
	fx.Provide(NewBotFromConfig),
	fx.Invoke(func(lc fx.Lifecycle, bot *Bot, logger logger.Logger) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := bot.Start(ctx); err != nil {
						logger.Errorf("Bot error: %v", err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				bot.Stop()
				return nil
			},
		})
	}),
)
