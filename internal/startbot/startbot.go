package startbot

import (
	"crypto/sha256"
	"encoding/hex"
	"os"

	"workwavebot/internal/logger"
	"workwavebot/internal/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func Createbot() (*tgbotapi.BotAPI, error) {

	// Получаем токен из переменных окружения
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		logger.ErrLog.Panic("Ошибка: токен бота не установлен! \n")
	}

	// Хэширование токена для логирования (не используется в NewBotAPI)
	hashedToken := hashToken(token)
	logger.BotLog.Printf("Хэш токена для проверки: %s \n", hashedToken)

	// Создаём объект бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.ErrLog.Fatalf("Ошибка создания бота: %v \n", err)
	}

	// Включаем отладочный режим
	bot.Debug = true

	return bot, nil
}

// запуск бота
func StartBot(b *telegram.Bot) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.GetUpdatesChan(u) // ← Bot сам умеет отдавать канал обновлений

	for update := range updates {

		// 1️⃣ обычные сообщения
		if update.Message != nil {
			b.HandleMessage(update.Message)
		}

		// 2️⃣ нажатия inline-кнопок
		if update.CallbackQuery != nil {
			b.HandleCallback(update.CallbackQuery)
		}
	}
}
