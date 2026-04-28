package telegram // файл state - это stateManager, отслеживание состояний

//делаем отдельные типы, а не просто string
type Menu string
type Action string

//         компилятор различает:
//state.Menu = MenuClients     // ✅
//state.Menu = ActionWaitName  // ❌ ошибка компиляции

// State Machine
type UserState struct {
	StackMenu          []Menu // меню в виде стека (последний зашёл - первый вышел)
	Action             Action // add_client, wait_percent, wait_admin_id и т.д
	PendingClientID    int    // ID клиента выбранного из списка
	PendingFormulaType string // тип формулы выбранной кнопкой
}

// перейти в новое меню — кладём в стек
func (s *UserState) PushMenu(menu Menu) {
	s.StackMenu = append(s.StackMenu, menu)
}

// вернуться назад — вытаскиваем из стека
func (s *UserState) PopMenu() Menu {
	if len(s.StackMenu) <= 1 {
		return MainMenu // уже в главном, некуда возвращаться
	}
	// убираем последний элемент
	s.StackMenu = s.StackMenu[:len(s.StackMenu)-1]
	// возвращаем новый последний — это и есть предыдущее меню
	return s.StackMenu[len(s.StackMenu)-1]
}

// текущее меню
func (s *UserState) CurrentMenu() Menu {
	if len(s.StackMenu) == 0 {
		return MainMenu
	}
	return s.StackMenu[len(s.StackMenu)-1]
}

// Сброс состояния — возвращаемся в главное меню и очищаем все данные
func (s *UserState) Reset() {
	s.StackMenu = []Menu{MainMenu}
	s.Action = ActionNone
	s.PendingClientID = 0
	s.PendingFormulaType = ""
}

//chatID  (ключ) → состояние пользователя (значение)
var userStates = make(map[int64]*UserState) // Карта для отслеживания состояний пользователей

// helper
func getUserState(chatID int64) *UserState {
	if userStates[chatID] == nil {
		s := &UserState{}
		s.Reset() // Reset кладёт MainMenu в стек и обнуляет всё остальное
		userStates[chatID] = s
	}
	return userStates[chatID]
}

// Менюшки
const (
	MainMenu Menu = "main" // первоначальное (главное) меню

	//меню рекрутера
	RecruiterMenu  Menu = "recruiter_menu"
	BonusMenu      Menu = "bonus_menu"
	UsefulnessMenu Menu = "usefulness_menu"

	//меню админа
	AdminMenu      Menu = "admin_menu"
	ClientsControl Menu = "clients_control"
	AllClients     Menu = "all_clients"

	AdminsControl Menu = "admins_control"
)

// Действия
const (
	ActionNone Action = "" // отсутствие действий
	// вернуться назад нет, т.к. это операция над стеком

	ActionWaitSalary Action = "wait_salary" // ждём зарплату от рекрутера

	// добавление нового клиента
	ActionAddClient      Action = "add_client"
	ActionWaitClientName Action = "wait_client_name"

	// привязка/обновление формулы
	ActionWaitFormulaType  Action = "wait_formula_type"  // ждём выбора кнопки
	ActionWaitFormulaValue Action = "wait_formula_value" // ждём числа текстом

	// админы
	ActionAddAdmin      Action = "add_admin"
	ActionWaitAdminID   Action = "wait_admin_id"
	ActionWaitAdminName Action = "wait_admin_name"
	ActionDeleteAdmin   Action = "delete_admin"

	// удаление
	ActionWaitDeleteConfirm Action = "wait_delete_confirm"
)
