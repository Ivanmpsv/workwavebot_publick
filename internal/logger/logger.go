package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	BotLog *log.Logger // для общих сообщений
	ErrLog *log.Logger // только для ошибок
)

func Init() error {

	// Гарантируем докеру, что папка logs существует
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		return fmt.Errorf("cannot create logs dir: %w", err)
	}

	// открываем файл, если его нет, то он создастся
	botLogFile, err := os.OpenFile("logs/botLog.log", os.O_CREATE|os.O_RDWR|
		os.O_APPEND, 0644) //O_RDWR для чиения и запси, 0644 - чтение всем, запись владельцу
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	// куда будут записываться логи | префикс для каждого сообщения | флаги формата логов
	BotLog = log.New(botLogFile, "BOT: ", log.Ldate|log.Ltime|log.Lshortfile)

	//// поскольку мы создали кастомные логи через log. new, то log.SetOutput нам не нужен

	//аналогично для лог-файла с ошибками
	errLogFile, err := os.OpenFile("logs/errLog.log", os.O_CREATE|os.O_RDWR|
		os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	ErrLog = log.New(errLogFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Устанавливаем файл как вывод для логов
	log.SetOutput(botLogFile)

	return nil
}
