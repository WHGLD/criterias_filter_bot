package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// GetToken возвращает токен бота из переменной окружения
func GetToken() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла | ", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}
	return token
}
