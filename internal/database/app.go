package database

import "database/sql"

//структура хранит указатель на подключение к базе данных
//Иначе во многих местах кода придётся подключатся к БД - sql.Open("postgres", connDB)
type App struct {
	db *sql.DB
}
