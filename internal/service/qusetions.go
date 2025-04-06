package service

// Questions Список вопросов
var Questions = []Question{
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
							Text: "HER2 pos.",
							Data: "q1_1_option2",
							NextQuestion: &Question{
								ID:   "q1_1_1",
								Text: "Выберите линию терапии:",
								Options: []Option{
									{
										Text:   "1 линия терапии",
										Data:   "q1_1_1_option1",
										Result: "BCD-267-1",
									},
									{
										Text:   "2 и последующие линии терапии",
										Data:   "q1_1_1_option2",
										Result: "CL011101223 (Перьета Р-фарм)",
									},
								},
							},
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
							Result: "Generium",
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
