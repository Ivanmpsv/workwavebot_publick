package database

import (
	"database/sql"
	"fmt"
	"os"
	"workwavebot/internal/logger"

	_ "github.com/lib/pq" // импорт драйвера, работает через database/sql - не пакет, отсюда _
)

func (app *App) ConnectDB() error {
	// читаем из переменных окружения — те же, что уже есть в .env
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")

	// шаг 1: создаём БД ww_bot если её нет
	if err := createDatabaseIfNotExists(user, password, host); err != nil {
		return fmt.Errorf("ensure database: %w", err)
	}

	// шаг 2: подключаемся уже к ww_bot
	connDB := fmt.Sprintf(
		"user=%s password=%s host=%s dbname=ww_bot sslmode=disable",
		user, password, host,
	)

	// sql.Open создаёт объект подключения к базе данных.
	// ❗️Важно: он НЕ устанавливает реальное соединение сразу.
	// Он просто подготавливает доступ к базе — «дескриптор подключения».
	db, err := sql.Open("postgres", connDB) //тут опять-таки стандартизировано "postgres"
	if err != nil {
		logger.ErrLog.Printf("ошибка при вызове дискроптора подключения %v \n", err)
		return fmt.Errorf("open db: %w", err)
	}

	// Теперь проверим реальное соединение с базой (ping).
	// Именно здесь происходит реальная попытка подключения.
	err = db.Ping()
	if err != nil {
		logger.ErrLog.Printf("не удалось подключиться к базе данных: %v \n", err)
		return fmt.Errorf("ping db: %w", err)
	}

	fmt.Println("Соединение: успешно")

	//сохраняем значение в экземпляр структуры App
	app.db = db

	return nil
}
