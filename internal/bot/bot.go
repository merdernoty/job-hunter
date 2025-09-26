package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/merdernoty/job-hunter/pkg/logger"
)

type Bot struct {
	api        *tgbotapi.BotAPI
	logger     logger.Logger
	webAppURL  string
	cancelFunc context.CancelFunc
}

func NewBot(token, webAppURL string, logger logger.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = false
	logger.Infof("Bot authorized on account %s", api.Self.UserName)

	return &Bot{
		api:       api,
		logger:    logger,
		webAppURL: webAppURL,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting Telegram bot...")

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Bot context cancelled, stopping...")
			return nil
		default:
			b.runUpdatesLoop(ctx)
		}
	}
}

func (b *Bot) runUpdatesLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Errorf("Bot updates loop panic: %v", r)
		}
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Stopping Telegram bot updates loop...")
			b.api.StopReceivingUpdates()
			return

		case update := <-updates:
			if update.UpdateID == 0 {
				continue
			}
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
	case "app":
		b.handleApp(msg)
	case "help":
		b.handleHelp(msg)
	default:
		if msg.Text != "" {
			b.handleDefault(msg)
		}
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	firstName := msg.From.FirstName
	if firstName == "" {
		firstName = "пользователь"
	}

	text := fmt.Sprintf(
		"👋 Привет, <b>%s</b>!\n\n"+
			"Добро пожаловать в <b>Job Hunter</b>!\n\n"+
			"🎯 Здесь ты можешь:\n"+
			"• Найти работу своей мечты\n"+
			"• Разместить вакансию\n"+
			"• Создать резюме\n"+
			"• Откликнуться на вакансии\n\n"+
			"🔗 Приложение: %s\n\n"+
			"Используй команду /app для получения ссылки",
		firstName,
		b.webAppURL,
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeHTML

	if _, err := b.api.Send(reply); err != nil {
		b.logger.Errorf("Failed to send start message: %v", err)
	}
}

func (b *Bot) handleApp(msg *tgbotapi.Message) {
	text := fmt.Sprintf("🔥 <b>Job Hunter Web App</b>\n\nПриложение: %s", b.webAppURL)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeHTML

	if _, err := b.api.Send(reply); err != nil {
		b.logger.Errorf("Failed to send app message: %v", err)
	}
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	text := `<b>🤖 Job Hunter Bot</b>

<b>Доступные команды:</b>
/start - Приветствие и ссылка на приложение
/app - Получить ссылку на приложение
/help - Показать эту справку

<b>О боте:</b>
Этот бот предоставляет доступ к платформе Job Hunter,
где вы можете найти работу или разместить вакансию.

<b>Приложение:</b>
` + b.webAppURL

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeHTML

	if _, err := b.api.Send(reply); err != nil {
		b.logger.Errorf("Failed to send help message: %v", err)
	}
}

func (b *Bot) handleDefault(msg *tgbotapi.Message) {
	text := fmt.Sprintf("Для работы с Job Hunter перейди по ссылке:\n%s", b.webAppURL)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(
				"📱 Открыть приложение",
				b.webAppURL,
			),
		),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyMarkup = keyboard

	if _, err := b.api.Send(reply); err != nil {
		b.logger.Errorf("Failed to send default message: %v", err)
	}
}

func (b *Bot) Stop() {
	b.logger.Info("Telegram bot stopped")
	b.api.StopReceivingUpdates()
}
