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

// todo проверка lastMessageID (с иммитацией кнопок назад + несколько пользаокв (не обязательно конкурентно))
