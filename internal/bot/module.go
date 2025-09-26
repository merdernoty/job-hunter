package bot

import (
	"context"

	"github.com/merdernoty/job-hunter/config"
	"github.com/merdernoty/job-hunter/pkg/logger"
	"go.uber.org/fx"
)

func NewBotFromConfig(cfg *config.Config, logger logger.Logger) (*Bot, error) {
	if cfg.Bot.Token == "" {
		logger.Warn("Bot token is not configured, skipping bot initialization")
		return nil, nil
	}

	if cfg.Bot.WebAppURL == "" {
		logger.Warn("WebApp URL is not configured")
	}

	return NewBot(cfg.Bot.Token, cfg.Bot.WebAppURL, logger)
}

var Module = fx.Module("bot",
	fx.Provide(NewBotFromConfig),
	fx.Invoke(func(lc fx.Lifecycle, bot *Bot, logger logger.Logger) {
		if bot == nil {
			logger.Info("Bot is not configured, skipping startup")
			return
		}

		logger.Info("Bot configured successfully, setting up lifecycle hooks")

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Info("Starting Telegram bot lifecycle hook...")
				botCtx, botCancel := context.WithCancel(context.Background())

				go func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Errorf("Bot goroutine panic: %v", r)
						}
					}()

					logger.Info("Bot goroutine starting...")
					if err := bot.Start(botCtx); err != nil {
						logger.Errorf("Bot error: %v", err)
					}
					logger.Info("Bot goroutine finished")
				}()
				bot.cancelFunc = botCancel

				logger.Info("Telegram bot started successfully")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info("Stopping Telegram bot lifecycle hook...")
				if bot.cancelFunc != nil {
					bot.cancelFunc()
				}
				bot.Stop()
				logger.Info("Telegram bot stopped successfully")
				return nil
			},
		})
	}),
)
