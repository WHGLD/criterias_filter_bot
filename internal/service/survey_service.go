package service

import (
	"errors"
	"sync"
)

// SurveyService Структура синглтон для работы с опросником
type SurveyService struct {
	mu               sync.RWMutex
	userAnswersMap   map[int64]*userAnswers
	lastMessageIDMap map[int64]int
}

// userAnswers Структура для хранения ответов пользователя
type userAnswers struct {
	currentQuestion *Question
	questionStack   []*Question
}

func (s *SurveyService) Start(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userAnswersMap[userID] = &userAnswers{
		currentQuestion: &Questions[0],
		questionStack:   []*Question{},
	}
}

func (s *SurveyService) Reset(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.userAnswersMap, userID)
}

func (s *SurveyService) PopFromQuestionStack(userID int64) (prevQuestion *Question, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userAnswersMap, ok := s.userAnswersMap[userID]
	if !ok {
		err = errors.New("USER STATE NOT FOUND IN MAP")
		return
	}

	stackLen := len(userAnswersMap.questionStack)
	if stackLen == 0 {
		err = errors.New("QUESTION STACK IS EMPTY")
		return
	}

	prevQuestion = userAnswersMap.questionStack[stackLen-1]
	s.userAnswersMap[userID].questionStack = userAnswersMap.questionStack[:stackLen-1]

	return
}

func (s *SurveyService) GetQuestionsStack(userID int64) (stack []*Question) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if mapByID, ok := s.userAnswersMap[userID]; ok {
		stack = mapByID.questionStack
	}
	return
}

func (s *SurveyService) SaveQuestionToStack(userID int64, question *Question) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.userAnswersMap[userID]; !ok {
		err = errors.New("USER STATE NOT FOUND IN MAP")
		return
	}

	s.userAnswersMap[userID].questionStack = append(s.userAnswersMap[userID].questionStack, question)
	return
}

func (s *SurveyService) GetCurrentQuestion(userID int64) (question *Question) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if mapByID, ok := s.userAnswersMap[userID]; ok {
		question = mapByID.currentQuestion
	}
	return
}

func (s *SurveyService) SetCurrentQuestion(userID int64, question *Question) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.userAnswersMap[userID]; !ok {
		err = errors.New("USER STATE NOT FOUND IN MAP")
		return
	}

	s.userAnswersMap[userID].currentQuestion = question
	return
}

func (s *SurveyService) GetLastMessageID(userID int64) (lastMessageID int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if mapByUserID, ok := s.lastMessageIDMap[userID]; ok {
		lastMessageID = mapByUserID
	}
	return
}

func (s *SurveyService) SetLastMessageID(userID int64, lastMessageID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastMessageIDMap[userID] = lastMessageID
	return
}

var (
	instance *SurveyService
	once     sync.Once
)

// GetInstance возвращает единственный экземпляр SurveyManager
func GetInstance() *SurveyService {
	once.Do(func() {
		instance = &SurveyService{
			userAnswersMap:   make(map[int64]*userAnswers),
			lastMessageIDMap: make(map[int64]int),
		}
	})
	return instance
}
