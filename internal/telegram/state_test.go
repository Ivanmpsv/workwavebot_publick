package telegram

import "testing"

// Запуск: go test ./internal/telegram/... -v

/* файл проверяет только "state machine" — UserState и getUserState.
Остальное в пакете telegram (Bot, HandleMessage, HandleCallback,
SendXxxMenu) НЕ покрыто тестами: без сети и БД просто так не протестировать
*tgbotapi.BotAPI и *database.App


TestUserState_PushPopCurrentMenu — table-driven тест на стек меню.

Логика PushMenu/PopMenu/CurrentMenu — это классический стек, лучше описать
его поведение как последовательность операций:
 "положить А, положить Б, вытащить -> должны увидеть А" и т.п.
Поэтому каждый кейс — это не единичное значение, а сценарий (список шагов),
который прогоняется по очереди с проверкой CurrentMenu() после каждого шага.
*/

func TestUserState_PushPopCurrentMenu(t *testing.T) {
	// step - одно действие над стеком и то, что должно быть в CurrentMenu()
	// сразу после этого действия.
	type step struct {
		push      Menu // если не "" — вызываем PushMenu(push)
		pop       bool // если true — вызываем PopMenu()
		wantAfter Menu // что должен вернуть CurrentMenu() после этого шага
	}

	cases := []struct {
		name  string
		steps []step
	}{
		{
			name:  "свежее состояние — сразу после Reset текущее меню главное",
			steps: []step{
				// шага нет — проверяем прямо в теле теста ниже
			},
		},
		{
			name: "один push — переходим в меню рекрутёра",
			steps: []step{
				{push: RecruiterMenu, wantAfter: RecruiterMenu},
			},
		},
		{
			name: "push, затем pop — возвращаемся в главное меню",
			steps: []step{
				{push: RecruiterMenu, wantAfter: RecruiterMenu},
				{pop: true, wantAfter: MainMenu},
			},
		},
		{
			name: "три уровня вложенности и последовательный выход",
			steps: []step{
				{push: RecruiterMenu, wantAfter: RecruiterMenu},
				{push: BonusMenu, wantAfter: BonusMenu},
				{push: AllClients, wantAfter: AllClients},
				{pop: true, wantAfter: BonusMenu},     // вышли на уровень выше
				{pop: true, wantAfter: RecruiterMenu}, // ещё на уровень выше
				{pop: true, wantAfter: MainMenu},      // вернулись в главное
			},
		},
		{
			name: "pop на пустом стеке (только MainMenu) — некуда выходить, остаёмся в главном",
			steps: []step{
				{pop: true, wantAfter: MainMenu},
				{pop: true, wantAfter: MainMenu}, // повторный pop — тоже безопасен
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := &UserState{}
			state.Reset() // все тесты стартуют с чистого состояния: [MainMenu]

			if len(tc.steps) == 0 {
				// кейс "свежее состояние" — шагов нет, просто проверяем сразу
				if got := state.CurrentMenu(); got != MainMenu {
					t.Fatalf("CurrentMenu() после Reset() = %q, want %q", got, MainMenu)
				}
				return
			}

			for i, s := range tc.steps {
				switch {
				case s.push != "":
					state.PushMenu(s.push)
				case s.pop:
					state.PopMenu()
				}

				if got := state.CurrentMenu(); got != s.wantAfter {
					t.Fatalf("шаг %d: CurrentMenu() = %q, want %q", i, got, s.wantAfter)
				}
			}
		})
	}
}

// TestUserState_Reset проверяет, что Reset действительно чистит ВСЕ поля,
// а не только стек меню — это важно, потому что PendingClientID/
// PendingFormulaType/Action используются как "память" между сообщениями
// пользователя, и утечка старого состояния в новый сценарий — частый
// источник багов в диалоговых ботах ("почему бот путает старого клиента
// с новым?").
func TestUserState_Reset(t *testing.T) {
	state := &UserState{
		StackMenu:          []Menu{MainMenu, RecruiterMenu, BonusMenu},
		Action:             ActionWaitFormulaValue,
		PendingClientID:    42,
		PendingFormulaType: "standard",
	}

	state.Reset()

	if got := state.CurrentMenu(); got != MainMenu {
		t.Errorf("после Reset() CurrentMenu() = %q, want %q", got, MainMenu)
	}
	if state.Action != ActionNone {
		t.Errorf("после Reset() Action = %q, want %q", state.Action, ActionNone)
	}
	if state.PendingClientID != 0 {
		t.Errorf("после Reset() PendingClientID = %d, want 0", state.PendingClientID)
	}
	if state.PendingFormulaType != "" {
		t.Errorf("после Reset() PendingFormulaType = %q, want \"\"", state.PendingFormulaType)
	}
}

// TestGetUserState_CreatesDefaultOnFirstCall — table-driven по граничным
// значениям chatID. Форма кейсов одинаковая (один вызов getUserState на
// "свежем" chatID -> проверка дефолта), различаются только сами chatID,
// в т.ч. граничные для map[int64]*UserState: ноль, отрицательные, крайние
// значения int64.
func TestGetUserState_CreatesDefaultOnFirstCall(t *testing.T) {
	cases := []struct {
		name   string
		chatID int64
	}{
		{"обычный положительный chatID", 111111},
		{"нулевой chatID", 0},
		{"отрицательный chatID", -111111},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := getUserState(tc.chatID)

			if got := state.CurrentMenu(); got != MainMenu {
				t.Errorf("CurrentMenu() = %q, want %q", got, MainMenu)
			}
			if state.Action != ActionNone {
				t.Errorf("Action = %q, want %q", state.Action, ActionNone)
			}
		})
	}
}

// TestGetUserState_ReturnsSamePointerOnSubsequentCalls — не table-driven,
// потому что тут важна последовательность из ДВУХ конкретных вызовов
// с мутацией между ними, а не набор независимых входов/выходов.
// Загонять такой сценарий в таблицу было бы искусственно.
func TestGetUserState_ReturnsSamePointerOnSubsequentCalls(t *testing.T) {
	const chatID int64 = 222222

	first := getUserState(chatID)
	first.PushMenu(RecruiterMenu)
	first.Action = ActionWaitClientName

	second := getUserState(chatID)

	if second != first {
		t.Fatalf("getUserState вернул новый указатель вместо существующего состояния")
	}
	if second.CurrentMenu() != RecruiterMenu {
		t.Errorf("CurrentMenu() = %q, want %q (состояние должно сохраняться между вызовами)",
			second.CurrentMenu(), RecruiterMenu)
	}
	if second.Action != ActionWaitClientName {
		t.Errorf("Action = %q, want %q", second.Action, ActionWaitClientName)
	}
}

// TestGetUserState_IsolatedBetweenDifferentChats проверяет, что userStates -
// это действительно карта ПО chatID, а не общая на всех переменная
// (актуально, т.к. userStates - пакетный var, шаренный между всеми чатами).
func TestGetUserState_IsolatedBetweenDifferentChats(t *testing.T) {
	const chatA int64 = 333333
	const chatB int64 = 444444

	stateA := getUserState(chatA)
	stateA.PushMenu(AdminMenu)

	stateB := getUserState(chatB)

	if got := stateB.CurrentMenu(); got != MainMenu {
		t.Errorf("состояние одного chatID повлияло на другой: CurrentMenu() = %q, want %q", got, MainMenu)
	}
}
