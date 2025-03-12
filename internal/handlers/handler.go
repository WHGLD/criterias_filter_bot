package handlers

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UserAnswers Структура для хранения ответов пользователя
type UserAnswers struct {
	Answers         map[string]string // questionID -> answer
	CurrentQuestion *Question         // Текущий активный вопрос
	QuestionStack   []*Question       // Стек пройденных вопросов
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

// Хранилище ответов пользователей
var userAnswersMap = make(map[int64]*UserAnswers)

// Мапа для быстрого доступа к вопросам
var questionsMap = make(map[string]*Question)

// ID последнего сообщения с вопросом
var lastMessageID int

// Список вопросов
var questions = []Question{
	{
		ID:   "q1",
		Text: "Выберите нозологию",
		Options: []Option{
			{
				Text: "Рак молочной железы",
				Data: "q1_option1",
				NextQuestion: &Question{
					ID:   "q1_1",
					Text: "Выберите подтип:",
					Options: []Option{
						{
							Text:   "Трижды негативный",
							Data:   "q1_1_option1",
							Result: "AREAL",
						},
						{
							Text:   "HER2 pos.",
							Data:   "q1_1_option2",
							Result: "CL011101223 (Перьета Р-фарм)",
						},
						{
							Text:   "HER2 pos. ИГХ 3+/2+",
							Data:   "q1_1_option3",
							Result: "BCD-267-1 (Энхерту)",
						},
					},
				},
			},
			{
				Text: "Колоректальный рак",
				Data: "q1_option2",
				NextQuestion: &Question{
					ID:   "q2_1",
					Text: "Какая предстоит линия лечения?",
					Options: []Option{
						{
							Text:   "1 линия",
							Data:   "q2_1_option1",
							Result: "CL01790199",
						},
						{
							Text:   "2 линия",
							Data:   "q2_1_option2",
							Result: "генериум",
						},
					},
				},
			},
			{
				Text: "Рак легкого",
				Data: "q1_option3",
				NextQuestion: &Question{
					ID:   "q3_1",
					Text: "Выберите молекулярно-генетический профиль:",
					Options: []Option{
						{
							Text:   "EGFR, ALK neg. PD-L >= 50%",
							Data:   "q3_1_option1",
							Result: "MIT-002",
						},
						{
							Text:   "EGFR, ALK neg. PD-L < 50%",
							Data:   "q3_1_option2",
							Result: "BEV-III/2022",
						},
					},
				},
			},
			{
				Text:   "Меланома",
				Data:   "q1_option4",
				Result: "MIT-002",
			},
			{
				Text:   "Рак головы и шеи",
				Data:   "q1_option5",
				Result: "р-фарм 2356",
			},
			{
				Text:   "Рак желудка",
				Data:   "q1_option6",
				Result: "р-фарм 1339",
			},
		},
	},
}

// Инициализация мапы вопросов
func init() {
	for i := range questions {
		registerQuestions(&questions[i])
	}
}

// Рекурсивная регистрация вопросов
func registerQuestions(q *Question) {
	questionsMap[q.ID] = q
	for _, opt := range q.Options {
		if opt.NextQuestion != nil {
			registerQuestions(opt.NextQuestion)
		}
	}
}

// HandleCallbackQuery Обработка нажатия Inline-кнопки
func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	userID := chatID

	if _, ok := userAnswersMap[userID]; !ok {
		userAnswersMap[userID] = &UserAnswers{
			Answers: make(map[string]string),
		}
	}

	userState := userAnswersMap[userID]

	if callbackQuery.Data == "start" {
		userAnswersMap[chatID] = &UserAnswers{
			Answers:         make(map[string]string),
			CurrentQuestion: &questions[0], // Начинаем с первого вопроса
		}

		fmt.Println("----")
		fmt.Println(lastMessageID)
		fmt.Println("----")

		editQuestion(bot, chatID, lastMessageID, &questions[0])
		return
	}

	if callbackQuery.Data == "back" {
		if len(userState.QuestionStack) > 0 {
			// Возвращаем предыдущий вопрос из стека
			prevQuestion := userState.QuestionStack[len(userState.QuestionStack)-1]
			userState.QuestionStack = userState.QuestionStack[:len(userState.QuestionStack)-1]
			userState.CurrentQuestion = prevQuestion
			editQuestion(bot, chatID, lastMessageID, prevQuestion)
		}
		return
	}

	// Поиск выбранной опции
	if userState.CurrentQuestion != nil {
		for _, option := range userState.CurrentQuestion.Options {
			if callbackQuery.Data == option.Data {
				userState.Answers[userState.CurrentQuestion.ID] = option.Text

				if option.Result != "" {
					sendResults(
						bot,
						lastMessageID,
						chatID,
						option.Data,
						option.Result,
						userState.Answers,
						false,
					)
					delete(userAnswersMap, userID)
					return
				}

				if option.NextQuestion != nil {
					// Сохраняем текущий вопрос в стек
					userState.QuestionStack = append(userState.QuestionStack, userState.CurrentQuestion)
					userState.CurrentQuestion = option.NextQuestion
					editQuestion(bot, chatID, lastMessageID, userState.CurrentQuestion)
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
		userAnswersMap[chatID] = &UserAnswers{
			Answers:         make(map[string]string),
			CurrentQuestion: &questions[0], // Начинаем с первого вопроса
		}
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
	if userState, ok := userAnswersMap[chatID]; ok && len(userState.QuestionStack) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "back"),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// sendResults отправка итогового результата с историей ответов
func sendResults(
	bot *tgbotapi.BotAPI,
	messageID int,
	chatID int64,
	resultOption string,
	result string,
	answers map[string]string,
	showAnswers bool,
) {
	messageText := ""

	if showAnswers {
		messageText += "📋 История ваших ответов:\n\n"

		for qID, answer := range answers {
			if q, ok := questionsMap[qID]; ok {
				messageText += fmt.Sprintf("▫️ *%s*\n➞ %s\n\n", q.Text, answer)
			}
		}
	}

	messageText += fmt.Sprintf("✅ *Подходящее исследование:* %s", result)
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
