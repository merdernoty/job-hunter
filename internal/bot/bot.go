package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/merdernoty/job-hunter/pkg/logger"
)

type Bot struct {
	api       *tgbotapi.BotAPI
	logger    logger.Logger
	webAppURL string
}

func NewBot(token, webAppURL string, logger logger.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = false
	logger.Infof("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:       api,
		logger:    logger,
		webAppURL: webAppURL,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting Telegram bot...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Stopping Telegram bot...")
			b.api.StopReceivingUpdates()
			return nil

		case update := <-updates:
			b.handleUpdate(update)
		}
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	b.logger.Infof("Received message from %s: %s", msg.From.UserName, msg.Text)

	switch msg.Command() {
	case "start":
		b.handleStart(msg)
	case "help":
		b.handleHelp(msg)
	default:
		b.sendWebAppLink(msg)
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	text := fmt.Sprintf(
		"Привет, %s! 👋\n\n"+
			"Добро пожаловать в Job Hunter!\n"+
			"Здесь ты можешь найти работу или разместить вакансию.\n\n"+
			"🔗 Ссылка на приложение: %s\n\n"+
			"Или используй команду /app",
		msg.From.FirstName,
		b.webAppURL,
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeHTML

	if _, err := b.api.Send(reply); err != nil {
		b.logger.Errorf("Failed to send start message: %v", err)
	}
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	text := `
<b>Job Hunter Bot</b>

<b>Команды:</b>
/start - Добро пожаловать
/app - Ссылка на приложение  
/help - Эта справка
`

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeHTML

	if _, err := b.api.Send(reply); err != nil {
		b.logger.Errorf("Failed to send help message: %v", err)
	}
}

func (b *Bot) sendWebAppLink(msg *tgbotapi.Message) {
	text := fmt.Sprintf("Job Hunter: %s", b.webAppURL)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)

	if _, err := b.api.Send(reply); err != nil {
		b.logger.Errorf("Failed to send webapp link: %v", err)
	}
}

func (b *Bot) Stop() {
	b.logger.Info("Telegram bot stopped")
	b.api.StopReceivingUpdates()
}
