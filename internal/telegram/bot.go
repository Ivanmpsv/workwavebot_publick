package telegram // центральная точка пакета — определяем структуру и конструктор.

import (
	"workwavebot/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api *tgbotapi.BotAPI
	app *database.App
}

func NewBot(api *tgbotapi.BotAPI, app *database.App) *Bot {
	return &Bot{api: api, app: app}
}

func (b *Bot) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return b.api.GetUpdatesChan(config)
}
