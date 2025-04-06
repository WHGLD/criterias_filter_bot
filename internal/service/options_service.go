package service

// Option Структура для варианта ответа
type Option struct {
	Text         string
	Data         string
	NextQuestion *Question // Следующий вопрос (если есть)
	Result       string    // Итоговый результат (если это конечный ответ)
}

func (o *Option) IsTerminal() bool {
	return o.Result != ""
}

func (o *Option) GetNextQuestion() *Question {
	return o.NextQuestion
}

func (o *Option) Matches(data string) bool {
	return o.Data == data
}
