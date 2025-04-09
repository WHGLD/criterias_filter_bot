package handlers

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"sync"
	"testing"

	"telegram-bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/mock"
)

type MockBot struct {
	mock.Mock
}

func (m *MockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	args := m.Called(c)
	return args.Get(0).(tgbotapi.Message), args.Error(1)
}

func (m *MockBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	args := m.Called(c)
	return args.Get(0).(*tgbotapi.APIResponse), args.Error(1)
}

func TestConcurrentUsers(t *testing.T) {
	var (
		userID, numUsers int
		mockBot          *MockBot
	)

	mockBot = new(MockBot)

	numUsers = 50
	var wg sync.WaitGroup

	expectedQuestion := service.Questions[0]

	for userID = 1; userID <= numUsers; userID++ {
		currentUserID := userID
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
			return msg.ChatID == int64(currentUserID) && msg.Text == expectedQuestion.Text
		})).Return(tgbotapi.Message{MessageID: currentUserID, Chat: &tgbotapi.Chat{ID: int64(currentUserID)}}, nil).Once()
	}

	wg.Add(numUsers)
	for userID = 1; userID <= numUsers; userID++ {
		go func(uID int) {
			defer wg.Done()
			msg := &tgbotapi.Message{
				Chat: &tgbotapi.Chat{
					ID: int64(uID),
				},
				Text: "/start@test_bot",
				Entities: []tgbotapi.MessageEntity{
					{
						Offset: 0,
						Length: 6,
						Type:   "bot_command",
					},
				},
			}
			HandleMessage(mockBot, msg)
		}(userID)
	}
	wg.Wait()

	mockBot.AssertNumberOfCalls(t, "Send", numUsers)
	// Проверяем, что все ожидаемые методы были вызваны
	mockBot.AssertExpectations(t)
}

// test case: ID q1 -> Data q1_option3 -> ID q3_1 -> Data q3_1_option1 -> result: MIT-002 (Рак легкого)
func TestSingleUserFlow(t *testing.T) {
	var (
		userID, messageID int
		surveyService     *service.SurveyService
		mockBot           *MockBot
		messageMock       tgbotapi.Message
	)

	mockBot = new(MockBot)
	surveyService = service.GetInstance()
	userID = 101
	messageID = 1 // постаянно его редактируем

	messageMock = tgbotapi.Message{MessageID: messageID, Chat: &tgbotapi.Chat{ID: int64(userID)}}

	mock.InOrder(
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
			return msg.Text == service.Questions[0].Text // Выберите нозологию
		})).Return(messageMock, nil).Once(),
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.EditMessageTextConfig) bool {
			return msg.Text == service.Questions[0].Options[2].NextQuestion.Text // Выберите молекулярно-генетический профиль:
		})).Return(messageMock, nil).Once(),
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.EditMessageTextConfig) bool {
			return strings.Contains(msg.Text, "Подходящее исследование")
		})).Return(messageMock, nil).Once(),
	)

	HandleMessage(mockBot, &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: int64(userID)},
		Text: "/start@test_bot",
		Entities: []tgbotapi.MessageEntity{
			{Type: "bot_command", Offset: 0, Length: 6},
		},
	})
	HandleCallbackQuery(mockBot, &tgbotapi.CallbackQuery{
		ID:      "callback_id",
		From:    &tgbotapi.User{ID: int64(userID)},
		Message: &messageMock,
		Data:    service.Questions[0].Options[2].Data, // q1_option3
	})

	expectedQuestion := service.Questions[0].Options[2].NextQuestion
	actualQuestion := surveyService.GetCurrentQuestion(int64(userID))
	assert.Equal(t, expectedQuestion, actualQuestion, "Ожидался следующий вопрос после выбора q1_option3")

	HandleCallbackQuery(mockBot, &tgbotapi.CallbackQuery{
		ID:      "final_callback",
		From:    &tgbotapi.User{ID: int64(userID)},
		Message: &messageMock,
		Data:    service.Questions[0].Options[2].NextQuestion.Options[0].Data, // q3_1_option1
	})

	// Проверяем, что все ожидаемые методы были вызваны
	mockBot.AssertExpectations(t)
	mockBot.AssertNumberOfCalls(t, "Send", 3)

	// Очищаем состояние после теста
	surveyService.Reset(int64(userID))
}

func TestSurveyWithStack(t *testing.T) {
	var (
		userID, messageID int
		surveyService     *service.SurveyService
		mockBot           *MockBot
		messageMock       tgbotapi.Message
	)

	mockBot = new(MockBot)
	surveyService = service.GetInstance()
	userID = 101
	messageID = 1 // постаянно его редактируем

	messageMock = tgbotapi.Message{MessageID: messageID, Chat: &tgbotapi.Chat{ID: int64(userID)}}

	mock.InOrder(
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
			return msg.Text == service.Questions[0].Text // Выберите нозологию
		})).Return(messageMock, nil).Once(),
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.EditMessageTextConfig) bool {
			return msg.Text == service.Questions[0].Options[0].NextQuestion.Text // Выберите подтип:
		})).Return(messageMock, nil).Once(),
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.EditMessageTextConfig) bool {
			return msg.Text == service.Questions[0].Options[0].NextQuestion.Options[1].NextQuestion.Text // Выберите линию терапии:
		})).Return(messageMock, nil).Once(),
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.EditMessageTextConfig) bool {
			return msg.Text == service.Questions[0].Options[0].NextQuestion.Text // Выберите подтип:
		})).Return(messageMock, nil).Once(),
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.EditMessageTextConfig) bool {
			return msg.Text == service.Questions[0].Text // Выберите нозологию
		})).Return(messageMock, nil).Once(),
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.EditMessageTextConfig) bool {
			return strings.Contains(msg.Text, "Подходящее исследование")
		})).Return(messageMock, nil).Once(),
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.EditMessageTextConfig) bool {
			return msg.Text == service.Questions[0].Text // Выберите нозологию
		})).Return(messageMock, nil).Once(),
	)

	HandleMessage(mockBot, &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: int64(userID)},
		Text: "/start@test_bot",
		Entities: []tgbotapi.MessageEntity{
			{Type: "bot_command", Offset: 0, Length: 6},
		},
	})

	// подмешаем парарельно еще 1 пользователя
	go imitateConcurrentUser(2, 102)

	HandleCallbackQuery(mockBot, &tgbotapi.CallbackQuery{
		ID:      "callback_id",
		From:    &tgbotapi.User{ID: int64(userID)},
		Message: &messageMock,
		Data:    service.Questions[0].Options[0].Data, // q1_option1
	})
	HandleCallbackQuery(mockBot, &tgbotapi.CallbackQuery{
		ID:      "callback_id",
		From:    &tgbotapi.User{ID: int64(userID)},
		Message: &messageMock,
		Data:    service.Questions[0].Options[0].NextQuestion.Options[1].Data, // q1_1_option2
	})

	// промежуточная проверка
	assertionsForStackTesting(
		t,
		messageID,
		int64(userID),
		2,
		service.Questions[0].Options[0].NextQuestion.Options[1].NextQuestion,
	)

	HandleCallbackQuery(mockBot, &tgbotapi.CallbackQuery{
		ID:      "back_callback",
		From:    &tgbotapi.User{ID: int64(userID)},
		Message: &messageMock,
		Data:    "back",
	})
	HandleCallbackQuery(mockBot, &tgbotapi.CallbackQuery{
		ID:      "back_callback",
		From:    &tgbotapi.User{ID: int64(userID)},
		Message: &messageMock,
		Data:    "back",
	})

	// подмешаем парарельно еще 1 пользователя
	go imitateConcurrentUser(3, 103)

	// промежуточная проверка
	assertionsForStackTesting(
		t,
		messageID,
		int64(userID),
		0,
		&service.Questions[0],
	)

	HandleCallbackQuery(mockBot, &tgbotapi.CallbackQuery{
		ID:      "final_callback",
		From:    &tgbotapi.User{ID: int64(userID)},
		Message: &messageMock,
		Data:    service.Questions[0].Options[5].Data, // q1_option6
	})
	HandleCallbackQuery(mockBot, &tgbotapi.CallbackQuery{
		ID:      "restart_callback",
		From:    &tgbotapi.User{ID: int64(userID)},
		Message: &messageMock,
		Data:    "start",
	})

	// подмешаем парарельно еще 1 пользователя
	go imitateConcurrentUser(4, 104)

	assertionsForStackTesting(
		t,
		messageID,
		int64(userID),
		0,
		&service.Questions[0],
	)

	mockBot.AssertExpectations(t)

	// Очищаем состояние после теста
	surveyService.Reset(int64(userID))
}

func imitateConcurrentUser(messageID, userID int) {
	var (
		mockBot     *MockBot
		messageMock tgbotapi.Message
	)

	mockBot = new(MockBot)
	messageMock = tgbotapi.Message{MessageID: messageID, Chat: &tgbotapi.Chat{ID: int64(userID)}}

	mock.InOrder(
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.MessageConfig) bool {
			return msg.Text == service.Questions[0].Text // Выберите нозологию
		})).Return(messageMock, nil).Once(),
		mockBot.On("Send", mock.MatchedBy(func(msg tgbotapi.EditMessageTextConfig) bool {
			return strings.Contains(msg.Text, "Подходящее исследование")
		})).Return(messageMock, nil).Once(),
	)

	HandleMessage(mockBot, &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: int64(userID)},
		Text: "/start@test_bot",
		Entities: []tgbotapi.MessageEntity{
			{Type: "bot_command", Offset: 0, Length: 6},
		},
	})
	HandleCallbackQuery(mockBot, &tgbotapi.CallbackQuery{
		ID:      "final_callback",
		From:    &tgbotapi.User{ID: int64(userID)},
		Message: &messageMock,
		Data:    service.Questions[0].Options[5].Data, // q1_option6
	})
}

func assertionsForStackTesting(
	t *testing.T,
	messageID int,
	userID int64,
	expectedStackLen int,
	expectedQuestion *service.Question,
) {
	surveyService := service.GetInstance()
	stackLen := len(surveyService.GetQuestionsStack(userID))
	assert.Equal(t, expectedStackLen, stackLen, "Проверка кол-во вопросов в стеке")

	actualQuestion := surveyService.GetCurrentQuestion(userID)
	assert.Equal(t, expectedQuestion, actualQuestion, "Ожидался другой следующий вопрос")

	lastMessageID := surveyService.GetLastMessageID(userID)
	assert.Equal(t, messageID, lastMessageID, "Проверка lastMessageID")
}
