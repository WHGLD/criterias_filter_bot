package handlers

import (
	"fmt"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UserAnswers Структура для хранения ответов пользователя
type UserAnswers struct {
	CurrentQuestion *Question
	QuestionStack   []*Question
}

// Question Структура для вопроса
type Question struct {
	ID      string
	Text    string
	Options []Option
}

// Option Структура для варианта ответа
type Option struct {
	Text         string
	Data         string
	NextQuestion *Question // Следующий вопрос (если есть)
	Result       string    // Итоговый результат (если это конечный ответ)
}

var (
	userAnswersMap = make(map[int64]*UserAnswers)
	lastMessageID  int
	mu             sync.RWMutex
)

// HandleCallbackQuery Обработка нажатия Inline-кнопки
func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	userID := chatID

	mu.Lock()
	if _, ok := userAnswersMap[userID]; !ok {
		userAnswersMap[userID] = &UserAnswers{}
	}
	mu.Unlock()

	if callbackQuery.Data == "start" {
		mu.Lock()
		userAnswersMap[chatID] = &UserAnswers{
			CurrentQuestion: &questions[0], // Начинаем с первого вопроса
		}
		mu.Unlock()

		editQuestion(bot, chatID, lastMessageID, &questions[0])
		return
	}

	if callbackQuery.Data == "back" {
		mu.Lock()
		if len(userAnswersMap[userID].QuestionStack) > 0 {
			// Возвращаем предыдущий вопрос из стека
			prevQuestion := userAnswersMap[userID].QuestionStack[len(userAnswersMap[userID].QuestionStack)-1]
			userAnswersMap[userID].QuestionStack = userAnswersMap[userID].QuestionStack[:len(userAnswersMap[userID].QuestionStack)-1]
			userAnswersMap[userID].CurrentQuestion = prevQuestion
			mu.Unlock()
			editQuestion(bot, chatID, lastMessageID, prevQuestion)
		} else {
			mu.Unlock()
		}
		return
	}

	if userAnswersMap[userID].CurrentQuestion != nil {
		for _, option := range userAnswersMap[userID].CurrentQuestion.Options {
			if callbackQuery.Data == option.Data {
				if option.Result != "" {
					sendResults(
						bot,
						lastMessageID,
						chatID,
						option.Data,
						option.Result,
					)
					mu.Lock()
					delete(userAnswersMap, userID)
					mu.Unlock()
					return
				}

				if option.NextQuestion != nil {
					// Сохраняем текущий вопрос в стек
					mu.Lock()
					userAnswersMap[userID].QuestionStack = append(userAnswersMap[userID].QuestionStack, userAnswersMap[userID].CurrentQuestion)
					userAnswersMap[userID].CurrentQuestion = option.NextQuestion
					mu.Unlock()
					editQuestion(bot, chatID, lastMessageID, userAnswersMap[userID].CurrentQuestion)
				}
				break
			}
		}
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Println(err)
	}
}

// HandleMessage Обработка текстового сообщения
func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	if message.Command() == "start" {
		mu.Lock()
		userAnswersMap[chatID] = &UserAnswers{
			CurrentQuestion: &questions[0], // Начинаем с первого вопроса
		}
		mu.Unlock()
		sendQuestion(bot, chatID, questions[0])
	}
}

// Универсальная функция для отправки вопроса
func sendQuestion(bot *tgbotapi.BotAPI, chatID int64, question Question) {
	keyboard := createKeyboard(&question, chatID)

	msg := tgbotapi.NewMessage(chatID, question.Text)
	msg.ReplyMarkup = keyboard
	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Println("Error sending message:", err)
		return
	}

	lastMessageID = sentMsg.MessageID
}

// Универсальная функция для редактирования вопроса
func editQuestion(bot *tgbotapi.BotAPI, chatID int64, messageID int, question *Question) {
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
func createKeyboard(question *Question, chatID int64) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Кнопки вариантов ответа
	for _, option := range question.Options {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(option.Text, option.Data),
		))
	}

	// Кнопка "Назад" если есть куда возвращаться
	mu.Lock()
	if _, ok := userAnswersMap[chatID]; ok && len(userAnswersMap[chatID].QuestionStack) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "back"),
		))
	}
	mu.Unlock()

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// sendResults отправка итогового результата с историей ответов
func sendResults(
	bot *tgbotapi.BotAPI,
	messageID int,
	chatID int64,
	resultOption string,
	result string,
) {
	messageText := fmt.Sprintf("✅ *Подходящее исследование:* %s", result)
	messageText += "\n\n" + responseDescriptions[resultOption]

	// Создаем inline-кнопку "Начать заново"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Начать заново", "start"),
		),
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		chatID,
		messageID,
		messageText,
		keyboard,
	)
	editMsg.ParseMode = "Markdown"

	if _, err := bot.Send(editMsg); err != nil {
		log.Println("Error sending results:", err)
	}
}
