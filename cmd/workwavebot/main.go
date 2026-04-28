package main

import (
	"log"
	"os"
	"workwavebot/internal/database"
	"workwavebot/internal/logger"
	"workwavebot/internal/startbot"
	"workwavebot/internal/telegram"

	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("BOT_TOKEN") == "" {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Ошибка загрузки .env файла: %v", err)
		}
	}

	// Инициализируем логгеры
	if err := logger.Init(); err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}

	//Подключаемся к базе данных PostgreSQL
	app := &database.App{}
	if err := app.ConnectDB(); err != nil {
		logger.ErrLog.Fatalf("Error connect DB: %v", err) // Fatalf, т.к нет смысла продолжать если нет подключения
	}
	if err := app.InitSchema(); err != nil {
		logger.ErrLog.Fatalf("Error init schema: %v", err)
	}

	// Создаём экземпляр бота
	api, err := startbot.Createbot()
	if err != nil {
		logger.ErrLog.Fatalf("%v", err)
	}

	b := telegram.NewBot(api, app) // ← собираем Bot из двух частей
	startbot.StartBot(b)           // ← передаём один объект вместо двух
}
