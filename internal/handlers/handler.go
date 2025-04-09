package handlers

import (
	"errors"
	"fmt"
	"log"
	"telegram-bot/internal/helper"

	"telegram-bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotInterface interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

// HandleCallbackQuery Обработка нажатия Inline-кнопки
func HandleCallbackQuery(bot BotInterface, callbackQuery *tgbotapi.CallbackQuery) {
	var (
		currentQuestion *service.Question
		option          service.Option
		chatID          int64
	)

	chatID = callbackQuery.Message.Chat.ID
	surveyService := service.GetInstance()

	if callbackQuery.Data == "start" {
		surveyService.Start(chatID)
		editQuestion(
			bot,
			chatID,
			surveyService.GetLastMessageID(chatID),
			surveyService.GetCurrentQuestion(chatID),
		)
		return
	}

	if callbackQuery.Data == "back" {
		if len(surveyService.GetQuestionsStack(chatID)) > 0 {
			prevQuestion, err := surveyService.PopFromQuestionStack(chatID)
			if err != nil {
				log.Println(err)
				return
			}

			if err = surveyService.SetCurrentQuestion(chatID, prevQuestion); err != nil {
				log.Println(err)
				return
			}

			editQuestion(bot, chatID, surveyService.GetLastMessageID(chatID), prevQuestion)
		}
		return
	}

	currentQuestion = surveyService.GetCurrentQuestion(chatID)
	if currentQuestion == nil {
		err := errors.New("currentQuestion == nil")
		log.Println(err)
		return
	}

	for _, option = range currentQuestion.Options {
		if !option.Matches(callbackQuery.Data) {
			continue
		}

		if option.IsTerminal() {
			sendResults(bot, surveyService.GetLastMessageID(chatID), chatID, option.Data, option.Result)
			surveyService.Reset(chatID)
			return
		}

		if nextQuestion := option.GetNextQuestion(); nextQuestion != nil {
			err := surveyService.SaveQuestionToStack(chatID, currentQuestion)
			if err != nil {
				log.Println(err)
				return
			}

			if err = surveyService.SetCurrentQuestion(chatID, nextQuestion); err != nil {
				log.Println(err)
				return
			}
			editQuestion(bot, chatID, surveyService.GetLastMessageID(chatID), nextQuestion)
			return
		}
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Println(err)
	}
}

// HandleMessage Обработка текстового сообщения
func HandleMessage(bot BotInterface, message *tgbotapi.Message) {
	if message.Command() == "start" {
		surveyService := service.GetInstance()
		surveyService.Start(message.Chat.ID)
		sendQuestion(bot, message.Chat.ID, *surveyService.GetCurrentQuestion(message.Chat.ID))
	}
}

// Универсальная функция для отправки вопроса
func sendQuestion(bot BotInterface, chatID int64, question service.Question) {
	keyboard := createKeyboard(&question, chatID)

	msg := tgbotapi.NewMessage(chatID, question.Text)
	msg.ReplyMarkup = keyboard
	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}

	surveyService := service.GetInstance()
	surveyService.SetLastMessageID(chatID, sentMsg.MessageID)
}

// Универсальная функция для редактирования вопроса
func editQuestion(bot BotInterface, chatID int64, messageID int, question *service.Question) {
	keyboard := createKeyboard(question, chatID)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		chatID,
		messageID,
		question.Text,
		keyboard,
	)

	if _, err := bot.Send(editMsg); err != nil {
		log.Println("Error editing message:", err)
	}
}

// Создание клавиатуры с кнопками
func createKeyboard(question *service.Question, chatID int64) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Кнопки вариантов ответа
	for _, option := range question.Options {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(option.Text, option.Data),
		))
	}

	// Кнопка "Назад" если есть куда возвращаться
	surveyService := service.GetInstance()
	currentQuestion := surveyService.GetCurrentQuestion(chatID)
	if currentQuestion != nil && len(surveyService.GetQuestionsStack(chatID)) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "back"),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Отправка итогового результата с историей ответов
func sendResults(
	bot BotInterface,
	messageID int,
	chatID int64,
	resultOption string,
	result string,
) {
	messageText := fmt.Sprintf("✅ *Подходящее исследование:* %s", result)
	messageText += "\n\n" + service.ResponseDescriptions[resultOption]

	// Создаем inline-кнопку "Начать заново"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Начать заново", "start"),
		),
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		chatID,
		messageID,
		helper.EscapeMarkdownV2(messageText),
		keyboard,
	)
	editMsg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(editMsg); err != nil {
		log.Println("Error sending results:", err)
	}
}
