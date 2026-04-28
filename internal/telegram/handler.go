package telegram

// файл handler.go — это обработчик сообщений и коллбеков от Telegram,
// тут логика что делать при каждом действии пользователя
// а также место соединения парсера и БД

import (
	"fmt"
	"strconv"
	"strings"
	"workwavebot/internal/parsers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- ОБРАБОТКА КНОПОК ---
func (b *Bot) HandleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	state := getUserState(chatID) // получаем состояние пользователя из карты по chatID, если его нет, то создаём новое (по логике функции getUserState)
	data := callback.Data         // данные, которые мы указали при создании кнопки (callbackData)

	// создаём ответ на коллбек, второй аргумент - текст уведомления, который может появиться у пользователя (необязательно)
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	b.api.Request(callbackConfig) // отправляем ответ на коллбек, чтобы Telegram знал, что мы обработали его (иначе может показывать "часики" у кнопки)

	// динамические коллбэки — обрабатываем до switch
	if strings.HasPrefix(data, "client:") {
		b.handleClientSelected(chatID, state, data)
		return
	}

	// подтверждение удаления клиента
	if strings.HasPrefix(data, "confirm:") {
		b.handleConfirm(chatID, state, data)
		return
	}

	switch Callback(data) {
	// ---Меню для рекрутера---
	case CallBackRecruierMenu:
		state.PushMenu(RecruiterMenu)
		state.Action = ActionNone
		b.SendRecruiterMenu(chatID)

	case CallbackBonus: // рассчитать бонус
		state.PushMenu(BonusMenu)
		state.Action = ActionNone
		b.SendBonusMenu(chatID)

	case CallbackUsefulness: // полезности
		state.PushMenu(UsefulnessMenu)
		state.Action = ActionNone
		b.SendUsefulness(chatID)

		// ---Меню для админа---
	case CallbackAdminMenu:
		state.PushMenu(AdminMenu)
		state.Action = ActionNone
		b.SendAdminMenu(chatID)

	case CallbackClientsControlMenu: // меню управления клиентами
		state.PushMenu(ClientsControl)
		state.Action = ActionNone
		b.SendClientsControlMenu(chatID)

	case CallbackAllClients:
		state.PushMenu(AllClients)
		state.Action = ActionNone
		b.sendClientListMenu(chatID, "Список всех клиентов:", "view")

	case CallbackAddClient: // добавить клиента
		state.Action = ActionAddClient
		b.startAddClient(chatID, state)
	case CallbackUpdateClient: // обновить формулу клиента
		state.Action = ActionWaitFormulaType
		b.startUpdateFormula(chatID, state)
	case CallbackDeleteClient: // удалить клиента
		state.Action = ActionWaitDeleteConfirm
		b.startDeleteClient(chatID, state)

	// ---выбор типа формулы---
	case CallbackFormulaStandard:
		b.receiveFormulaType(chatID, state, "standard")
	case CallbackFormulaSalary:
		b.receiveFormulaType(chatID, state, "salary")
	case CallbackFormulaFree:
		b.receiveFormulaType(chatID, state, "free")

	case CallbackAdminsControlMenu:
		state.PushMenu(AdminsControl) // добавь константу в state.go
		state.Action = ActionNone
		b.SendAdminsControlMenu(chatID)

	case CallbackAllAdmins:
		b.showAllAdmins(chatID)

	case CallbackAddAdmin:
		b.startAddAdmin(chatID, state)

	case CallbackDeleteAdmin:
		b.startDeleteAdmin(chatID, state)

	case CallbackBack:
		prev := state.PopMenu() // вытаскиваем предыдущее
		state.Action = ActionNone
		b.sendMenu(chatID, prev) // отправляем нужное меню
	}

}

// --- ОБРАБОТКА ТЕКСТОВЫХ СООБЩЕНИЙ ---
func (b *Bot) HandleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	state := getUserState(chatID)

	// активные ожидания проверяем первыми
	switch state.Action {
	case ActionWaitClientName:
		b.receiveClientName(chatID, state, message.Text)
		return
	case ActionWaitFormulaValue:
		b.receiveFormulaValue(chatID, state, message.Text)
		return
	case ActionWaitSalary:
		b.receiveSalary(chatID, state, message.Text)
		return

	case ActionWaitAdminID:
		b.receiveAdminInput(chatID, state, message.Text)
		return
	case ActionDeleteAdmin:
		b.receiveDeleteAdminInput(chatID, state, message.Text)
		return
	}

	switch message.Text {
	case "/start", "старт", "Старт", "привет", "Лок-тар", "лок-тар", "локтар", "Локтар":
		state.Reset() // очищаем стек, иначе будет несколько MainMenu в памяти стека
		state.PushMenu(MainMenu)
		state.Action = ActionNone
		b.SendMainMenu(chatID)
	}
}

// метод - вернуться в прошлое меню

func (b *Bot) sendMenu(chatID int64, menu Menu) {
	switch menu {
	case MainMenu:
		b.SendMainMenu(chatID)
	case AdminMenu:
		b.SendAdminMenu(chatID)
	case ClientsControl:
		b.SendClientsControlMenu(chatID)
	case RecruiterMenu:
		b.SendRecruiterMenu(chatID)
	case AdminsControl:
		b.SendAdminsControlMenu(chatID)
		// СЮДА НОВЫЕ МЕНЮШКИ (если будут)
	}
}

// --- добавление клиента ---

func (b *Bot) startAddClient(chatID int64, state *UserState) {

	state.Action = ActionWaitClientName
	b.SendAddClientNameMenu(chatID)
}

func (b *Bot) receiveClientName(chatID int64, state *UserState, name string) {
	name = strings.TrimSpace(name) // удаляет пробелы в наче и конце
	if name == "" {
		b.SendMessage(chatID, "Имя не может быть пустым, введите ещё раз:")
		return // не меняем state — ждём повтора
	}

	if err := b.app.AddClient(name); err != nil {
		b.SendMessage(chatID, "Ошибка при добавлении клиента")
		state.Reset()
		return
	}
	b.SendMessage(chatID, "Клиент \""+name+"\" успешно добавлен")
	state.Reset()
	b.SendClientsControlMenu(chatID)
}

// --- удаление клиента ---

func (b *Bot) startDeleteClient(chatID int64, state *UserState) {
	state.Action = ActionWaitDeleteConfirm
	b.SendChooseClientForDeleteMenu(chatID)
}

// --- обновление формулы ---

func (b *Bot) startUpdateFormula(chatID int64, state *UserState) {
	state.Action = ActionWaitFormulaType
	b.SendChooseClientForPercentMenu(chatID)
}

// пользователь выбрал клиента из списка - что, в каком случае делаем
func (b *Bot) handleClientSelected(chatID int64, state *UserState, data string) {
	// "client:update:42" → ["client", "update", "42"]
	parts := strings.SplitN(data, ":", 3)
	if len(parts) != 3 {
		return
	}

	context := parts[1] // "update" или "delete"
	clientID, err := strconv.Atoi(parts[2])
	if err != nil {
		b.SendMessage(chatID, "Некорректный ID клиента")
		return
	}

	switch context {
	case "bonus":
		b.startBonusCalc(chatID, state, clientID)

	case "view":
		state.PendingClientID = clientID
		state.Action = ActionNone
		b.SendBonusMenu(chatID) // TODO: сделать просмотр информации о клиенте, а не рассчёт бонуса

	case "update":
		// сохраняем ID, переходим к выбору типа формулы
		state.PendingClientID = clientID
		state.Action = ActionWaitFormulaType
		b.SendAddClientPercentMenu(chatID)

	case "delete":
		// сохраняем ID, просим подтверждение
		state.PendingClientID = clientID
		state.Action = ActionWaitDeleteConfirm
		b.sendDeleteConfirmMenu(chatID, clientID)
	}
}

// пользователь нажал кнопку типа формулы
func (b *Bot) receiveFormulaType(chatID int64, state *UserState, formulaType string) {
	if state.Action != ActionWaitFormulaType || state.PendingClientID == 0 {
		return
	}

	if formulaType == "free" {
		// free формула — вводится текстом, не числом
		state.PendingFormulaType = formulaType
		state.Action = ActionWaitFormulaValue
		b.SendMessage(chatID, "Введите формулу расчёта:")
		return
	}

	// standard и salary — вводится числом
	state.PendingFormulaType = formulaType
	state.Action = ActionWaitFormulaValue

	switch formulaType {
	case "standard":
		b.SendMessage(chatID, "Введите процент от годового дохода и нажмите enter (в формате: 0.15, 0.10 и т.п):")
	case "salary":
		b.SendMessage(chatID, "Введите число-коэффициент и нажмите enter\n\n"+
			"1 = один оклад гросс, 0.5 = половина оклада гросс и т.д.): ")
	}
}

// пользователь ввёл значение формулы текстом
func (b *Bot) receiveFormulaValue(chatID int64, state *UserState, input string) {
	clientID := state.PendingClientID
	formulaType := state.PendingFormulaType

	var saveErr error

	switch formulaType {
	case "standard":
		value, err := parsers.ParseFloat(input)
		if err != nil {
			b.SendMessage(chatID, "Введите число, например: 0.15")
			return // не меняем state — ждём повтора
		}
		// обновляем тип и значение
		saveErr = b.app.ChangeFormulaType(clientID, "standard")
		if saveErr == nil {
			saveErr = b.app.SetStandardFormula(clientID, value)
		}

	case "salary":
		value, err := parsers.ParseFloat(input)
		if err != nil {
			b.SendMessage(chatID, "Введите число, например: 1 или 0.5")
			return
		}
		saveErr = b.app.ChangeFormulaType(clientID, "salary")
		if saveErr == nil {
			saveErr = b.app.SetSalaryFormula(clientID, value)
		}

	case "free":
		text := strings.TrimSpace(input)
		if text == "" {
			b.SendMessage(chatID, "Формула не может быть пустой, введите ещё раз:")
			return
		}
		saveErr = b.app.ChangeFormulaType(clientID, "free")
		if saveErr == nil {
			saveErr = b.app.SetFreeFormula(clientID, text)
		}
	}

	if saveErr != nil {
		b.SendMessage(chatID, "Ошибка при сохранении, проверьте логи")
		state.Reset()
		return
	}

	b.SendMessage(chatID, "Формула успешно сохранена ✓")
	state.Reset()
	b.SendClientsControlMenu(chatID)
}

// подтверждение удаления клиента (показать сообщение)
func (b *Bot) sendDeleteConfirmMenu(chatID int64, clientID int) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Да, удалить", fmt.Sprintf("confirm:delete:%d", clientID),
			),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", string(CallbackBack)),
		),
	)
	msg := tgbotapi.NewMessage(chatID, "Вы уверены? Это действие необратимо.")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// обработка подтверждения удаления клиента (обработка sendDeleteConfirmMenu)
func (b *Bot) handleConfirm(chatID int64, state *UserState, data string) {
	// "confirm:delete:42" → ["confirm", "delete", "42"]
	parts := strings.SplitN(data, ":", 3)
	if len(parts) != 3 {
		return
	}

	action := parts[1]
	clientID, err := strconv.Atoi(parts[2])
	if err != nil {
		return
	}

	switch action {
	case "delete":
		if err := b.app.DELETEClient(clientID); err != nil {
			b.SendMessage(chatID, "Ошибка при удалении, проверьте логи")
		} else {
			b.SendMessage(chatID, "Клиент успешно удалён ✓")
		}
		state.Reset()
		b.SendClientsControlMenu(chatID)
	}
}

// получение формулы клиента
func (b *Bot) receiveSalary(chatID int64, state *UserState, input string) {
	salary, err := parsers.ParseFloat(input)
	if err != nil {
		b.SendMessage(chatID, "Введите число, например: 150000")
		return
	}

	formula, err := b.app.GetClientFormula(state.PendingClientID)
	if err != nil {
		b.SendMessage(chatID, "Ошибка: "+err.Error())
		state.Reset()           //сброс состояния
		b.SendBonusMenu(chatID) // вернуться к выбору клиента
		return
	}

	b.SendMessage(chatID, formula.Calculate(salary)) //  вызываем метод
	state.Reset()
	b.SendBonusMenu(chatID) // вернуться к выбору клиента
}

// сам подсчёт бонуса
func (b *Bot) startBonusCalc(chatID int64, state *UserState, clientID int) {
	state.PendingClientID = clientID
	state.Action = ActionWaitSalary
	b.SendMessage(chatID, "Введите зарплату кандидата (в гроссах):")

}

// --- ДЕЙСТВИЯ НАД АДМИНА ---

func (b *Bot) showAllAdmins(chatID int64) {
	admins, err := b.app.GetAdmins()
	if err != nil {
		b.SendMessage(chatID, "Ошибка при загрузке списка")
		return
	}
	if len(admins) == 0 {
		b.SendMessage(chatID, "Список админов пуст")
		return
	}

	var lines []string
	for _, ad := range admins {
		lines = append(lines, fmt.Sprintf("%d — %s", ad.ID, ad.Name))
	}
	b.SendMessage(chatID, strings.Join(lines, "\n"))
}

//добавление админа

func (b *Bot) startAddAdmin(chatID int64, state *UserState) {
	if !b.app.CheckAdmin(chatID) {
		b.SendMessage(chatID, "У вас нет прав для этого действия")
		return
	}
	state.Action = ActionWaitAdminID
	b.SendMessage(chatID, "Введите Telegram ID и имя нового админа через пробел:\nПример: 123456789 Иван")
}

func (b *Bot) receiveAdminInput(chatID int64, state *UserState, input string) {
	id, name, err := parsers.ParseAdminInput(input)
	if err != nil {
		b.SendMessage(chatID, "Неверный формат. Введите ID и имя через пробел:\nПример: 123456789 Иван")
		return // не сбрасываем state — ждём повтора
	}

	if err := b.app.AddAdmin(id, name); err != nil {
		b.SendMessage(chatID, "Ошибка при добавлении, проверьте логи")
		state.Reset()
		return
	}

	b.SendMessage(chatID, fmt.Sprintf("Админ %s успешно добавлен ✓", name))
	state.Reset()
	b.SendAdminsControlMenu(chatID)
}

// удаление админа

func (b *Bot) startDeleteAdmin(chatID int64, state *UserState) {
	if !b.app.CheckAdmin(chatID) {
		b.SendMessage(chatID, "У вас нет прав для этого действия")
		return
	}
	state.Action = ActionDeleteAdmin
	b.SendMessage(chatID, "Введите Telegram ID админа для удаления:")
}

func (b *Bot) receiveDeleteAdminInput(chatID int64, state *UserState, input string) {
	id, err := parsers.ParseInt64(input) // нужно добавить в parsers.go
	if err != nil {
		b.SendMessage(chatID, "Введите корректный Telegram ID, например: 123456789")
		return
	}

	if err := b.app.DeleteAdmin(id); err != nil {
		b.SendMessage(chatID, "Ошибка: "+err.Error())
		state.Reset()
		return
	}

	b.SendMessage(chatID, "Админ успешно удалён")
	state.Reset()
	b.SendAdminsControlMenu(chatID)
}
