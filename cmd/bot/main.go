package main

import (
	"log"

	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Получаем токен из переменной окружения
	token := config.GetToken()

	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false // Включаем отладку

	// Настраиваем канал обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.CallbackQuery != nil { // Если нажата Inline-кнопка
			handlers.HandleCallbackQuery(bot, update.CallbackQuery)
		} else if update.Message != nil { // Если есть новое сообщение
			handlers.HandleMessage(bot, update.Message)
		} else {
			log.Println("command not found: ", update)
		}
	}
}
