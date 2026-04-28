package database

import (
	"database/sql"
	"fmt"
)

// createDatabaseIfNotExists подключается к системной БД postgres
// и создаёт ww_bot, если её ещё нет.
// Нельзя создать БД находясь внутри неё — нужно отдельное соединение.
func createDatabaseIfNotExists(user, password, host string) error {
	// подключаемся к системной БД postgres, она есть всегда
	connSys := fmt.Sprintf(
		"user=%s password=%s host=%s dbname=postgres sslmode=disable",
		user, password, host,
	)

	sys, err := sql.Open("postgres", connSys)
	if err != nil {
		return fmt.Errorf("sys open: %w", err)
	}
	defer sys.Close()

	if err = sys.Ping(); err != nil {
		return fmt.Errorf("sys ping: %w", err)
	}

	// проверяем — существует ли уже ww_bot
	var exists bool
	err = sys.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM pg_database WHERE datname = 'ww_bot'
        )
    `).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check db exists: %w", err)
	}

	if !exists {
		// CREATE DATABASE нельзя выполнить внутри транзакции,
		// поэтому используем Exec напрямую — это нормально
		_, err = sys.Exec(`CREATE DATABASE ww_bot`)
		if err != nil {
			return fmt.Errorf("create database: %w", err)
		}
		fmt.Println("База данных ww_bot создана")
	}

	return nil
}

// InitSchema создаёт все таблицы и типы внутри ww_bot,
// если они ещё не существуют. Безопасно вызывать при каждом запуске.
func (app *App) InitSchema() error {
	queries := []string{

		// ENUM-тип нельзя создать через IF NOT EXISTS —
		// такого синтаксиса в PostgreSQL нет.
		// Используем DO-блок: пробуем создать, ловим ошибку если уже есть.
		`DO $$
        BEGIN
            CREATE TYPE formula_kind AS ENUM ('standard', 'salary', 'free');
        EXCEPTION
            WHEN duplicate_object THEN NULL; -- тип уже есть, просто идём дальше
        END
        $$`,

		// admins: id — это Telegram ID пользователя (int64), не автоинкремент
		`CREATE TABLE IF NOT EXISTS admins (
            id   BIGINT PRIMARY KEY,
            name TEXT
        )`,

		// clients: id — автоинкремент через SERIAL
		`CREATE TABLE IF NOT EXISTS clients (
            id           SERIAL PRIMARY KEY,
            name         TEXT,
            formula_type formula_kind
        )`,

		// три таблицы формул — у каждой client_id это одновременно:
		// PRIMARY KEY (одна формула на клиента)
		// FOREIGN KEY → clients.id с CASCADE (удалил клиента — формула уйдёт сама)
		`CREATE TABLE IF NOT EXISTS standard_formulas (
            client_id INTEGER PRIMARY KEY
                REFERENCES clients(id) ON DELETE CASCADE,
            percent   NUMERIC NOT NULL
        )`,

		`CREATE TABLE IF NOT EXISTS salary_formulas (
            client_id INTEGER PRIMARY KEY
                REFERENCES clients(id) ON DELETE CASCADE,
            coefficient    NUMERIC NOT NULL
        )`,

		`CREATE TABLE IF NOT EXISTS free_formulas (
            client_id INTEGER PRIMARY KEY
                REFERENCES clients(id) ON DELETE CASCADE,
            formula   TEXT NOT NULL
        )`,
	}

	for _, q := range queries {
		if _, err := app.db.Exec(q); err != nil {
			return fmt.Errorf("init schema: %w", err)
		}
	}

	fmt.Println("Схема БД готова")
	return nil
}
