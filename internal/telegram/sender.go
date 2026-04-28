package telegram //файл sender нужен для отправки сообщений и менб-кнопок

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// функция wrapper (обёртка) чтоб не дублировать код, создаёт и отправляет сообщение
func (b *Bot) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text) // функция создаёт новое сообщение
	b.api.Send(msg)                          // отправляет сообщение
}

// главное меню бота (всплывающие кнопки)
func (b *Bot) SendMainMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Локтар огар! Выберите действие: ")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню Рекрутера", string(CallBackRecruierMenu)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню Админа", string(CallbackAdminMenu)),
		),
	)

	msg.ReplyMarkup = keyboard // "прикрепить клавиатуру к сообщению"
	b.api.Send(msg)

}

// показать всех клиентов кнопками (используется в разных частях бота)
func (b *Bot) sendClientListMenu(chatID int64, prompt string, context string) {
	//prompt - текст-сообщение над кнопками (пояснение к выбору или действию)
	//contex - строка, которая будет частью callbackData, дабы понять в каком
	//контексте мы выбираем клиента (для расчёта бонуса или для обновления формулы)

	clients, err := b.app.GetClients()
	if err != nil {
		b.SendMessage(chatID, "Ошибка при загрузке клиентов")
		return
	}

	if len(clients) == 0 {
		b.SendMessage(chatID, "Клиентов пока нет")
		return
	}
	// для каждого клиента создаём кнопку, которая при нажатии будет отправлять callback с данными "client:{контекст}:{id}"
	var rows [][]tgbotapi.InlineKeyboardButton // двумерный слайс т.к. кнопки располагаются в строках, а строки в клавиатуре
	for _, c := range clients {
		callbackData := fmt.Sprintf("client:%s:%d", context, c.ID)
		btn := tgbotapi.NewInlineKeyboardButtonData(c.Name, callbackData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	// кнопка "Назад"
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", string(CallbackBack)),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, prompt)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

//     ------ МЕНЮ ДЛЯ РЕКРУТЕРОВ -------

func (b *Bot) SendRecruiterMenu(chatID int64) {

	msg := tgbotapi.NewMessage(chatID, "Что делаем?")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Рассчитать бонус", string(CallbackBonus)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Полезности", string(CallbackUsefulness)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", string(CallbackBack)),
		),
	)
	msg.ReplyMarkup = keyboard // "прикрепить клавиатуру к сообщению"
	b.api.Send(msg)
}

func (b *Bot) SendBonusMenu(chatID int64) {

	b.sendClientListMenu(chatID, "Выберите клиента для расчёта бонуса:", "bonus")

}

// Полезности (пока просто текст с ссылками)
func (b *Bot) SendUsefulness(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Полезности для рекрутера:\n\n"+
		"База знаний Week: https://app.weeek.net/ws/790970/kb/2\n\n"+
		"CRM, где ведём кандидатов: https://huntlee.ru/main/\n\n"+
		"Наш главный клиент - Мастерская: https://maya.it.ru/\n\n")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", string(CallbackBack)),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

//   --------- МЕНЮ ДЛЯ АДМИНОВ --------

func (b *Bot) SendAdminMenu(chatID int64) {

	if !b.app.CheckAdmin(chatID) {
		b.SendMessage(chatID, "У вас нет прав для этого действия")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите менюшку для дальнейших действий")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Управление клиентами", string(CallbackClientsControlMenu)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Управление админами", string(CallbackAdminsControlMenu)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", string(CallbackBack)),
		),
	)

	msg.ReplyMarkup = keyboard // "прикрепить клавиатуру к сообщению"
	b.api.Send(msg)
}

// --- Менюшки - работа с клиентами ---
func (b *Bot) SendClientsControlMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Выберите желаемое действие: ")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Показать список всех клиентов", string(CallbackAllClients)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить нового клиента", string(CallbackAddClient)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Обновить формулу клиента", string(CallbackUpdateClient)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить клиента", string(CallbackDeleteClient)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", string(CallbackBack)),
		),
	)

	msg.ReplyMarkup = keyboard // "прикрепить клавиатуру к сообщению"
	b.api.Send(msg)
}

// Добавить ИМЯ клиента (только имя, формулу потом)
func (b *Bot) SendAddClientNameMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Введите (вручную) имя клиента и нажмите enter:")
	b.api.Send(msg)
}

// Выбрать клиента для обновления формулы
func (b *Bot) SendChooseClientForPercentMenu(chatID int64) {
	b.sendClientListMenu(chatID, "Выберите клиента для обновления формулы:", "update")
}

// Добавить или обновить ФОРМУЛУ для существующего клиента в БД
func (b *Bot) SendAddClientPercentMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Выбрете формулу для расчёта бонуса, где:\n"+
		"1. Standard - процент от годового дохода \n"+
		"2. Salary - клиент оплачивает соразмерно окладу кандидата\n"+
		"3. Free - особый вид расчётов*\n\n"+

		"* для формулы Free используется внешняя ИИ, \n")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Standard", "formula:standard"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Salary", "formula:salary"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Free", "formula:free"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", string(CallbackBack)),
		),
	)

	msg.ReplyMarkup = keyboard // "прикрепить клавиатуру к сообщению"
	b.api.Send(msg)
}

// Удалить клиента (выбрать из списка)
func (b *Bot) SendChooseClientForDeleteMenu(chatID int64) {
	b.sendClientListMenu(chatID, "Выберите клиента для удаления:", "delete")
}

// --- Менюшки - работа с админами ---

func (b *Bot) SendAdminsControlMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Управление админами:")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Список всех админов", string(CallbackAllAdmins)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить админа", string(CallbackAddAdmin)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить админа", string(CallbackDeleteAdmin)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", string(CallbackBack)),
		),
	)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}
