package database // файл clients.go специализированные запросы или инфа по клиентам

import (
	"database/sql"
	"fmt"
	"workwavebot/internal/calculator"
	"workwavebot/internal/logger"
)

type Clients struct {
	ID          int
	Name        string
	FormulaType sql.NullString // Go не умеет преобразовать NULL в string

	/* sql.NullString — специальная обёртка из стандартной библиотеки
	    type NullString struct {
	        String string  -- само значение
	       Valid  bool    -- true если не NULL, false если NULL
	   }
	*/
}

// --- Структуры для чтения данных из БД ---
// стилистика наименования как в PostgreSQL
type Standard_formulas struct {
	ID      int
	Percent float64
}

type Salary_formulas struct {
	ID          int
	Coefficient float64
}

type Free_formulas struct {
	ID      int
	Formula string
}

// Интерфейс, который умеет считать результат
type Formula interface {
	Calculate(salary float64) string
}

// структуры для расчёта бонуса
type StandardFormula struct {
	Percent float64
}

type SalaryFormula struct {
	Coefficient float64
}

type FreeFormula struct {
	Text string
}

func (f StandardFormula) Calculate(salary float64) string {
	return calculator.StandartFormula(salary, f.Percent)
}

func (f SalaryFormula) Calculate(salary float64) string {
	return calculator.SalaryFormula(salary, f.Coefficient)
}

func (f FreeFormula) Calculate(salary float64) string {
	return calculator.FreeFormula(f.Text)
}

// === РАБОТА ДАННЫМИ О КЛИЕНТАХ ===

// добавить клиента
func (a *App) AddClient(name string) error {
	// сначало создаётся только имя, формула будет NULL

	// SQL-команда вставки — используем плейсхолдеры $1, $2, $3 ...
	query := `
	INSERT INTO clients (name)
	VALUES ($1)
	`

	// db.Exec выполняет INSERT-запрос без возврата данных
	_, err := a.db.Exec(query, name)
	if err != nil {
		logger.ErrLog.Printf("ошибка при заполнении таблицы: %v \n", err)

		return fmt.Errorf("AddClient: %w", err)
	}

	return nil
}

// добавить/изменить тип формулы клиента
func (a *App) ChangeFormulaType(clientID int, formulaType string) error {
	switch formulaType {
	case "standard", "salary", "free":
	default:
		return fmt.Errorf("неизвестный тип формулы: %s", formulaType)
	}

	query := `
	UPDATE clients
	SET formula_type = $2
	WHERE id = $1
	`
	_, err := a.db.Exec(query, clientID, formulaType)
	if err != nil {
		logger.ErrLog.Printf("ошибка ChangeFormulaType: %v", err)
		return fmt.Errorf("change formula type: %w", err)
	}
	return nil
}

// ------ТИПЫ ФОРМУЛ------

func (a *App) SetStandardFormula(clientID int, percent float64) error {
	query := `
	INSERT INTO standard_formulas (client_id, percent)
	VALUES ($1, $2)
	ON CONFLICT (client_id) DO UPDATE SET percent = $2
	`
	_, err := a.db.Exec(query, clientID, percent)
	if err != nil {
		logger.ErrLog.Printf("ошибка при SQL запросе INSERT: %v", err)
		return fmt.Errorf("insert standard formula: %w", err)
	}
	return nil
}

func (a *App) SetSalaryFormula(clientID int, coefficient float64) error {
	query := `
	INSERT INTO salary_formulas (client_id, coefficient)
	VALUES ($1, $2)
	ON CONFLICT (client_id) DO UPDATE SET coefficient = $2
	`
	_, err := a.db.Exec(query, clientID, coefficient)
	if err != nil {
		logger.ErrLog.Printf("ошибка при SQL запросе INSERT: %v", err)
		return fmt.Errorf("insert salary formula: %w", err)
	}
	return nil
}

func (a *App) SetFreeFormula(clientID int, formula string) error {
	query := `
	INSERT INTO free_formulas (client_id, formula)
	VALUES ($1, $2)
	ON CONFLICT (client_id) DO UPDATE SET formula = $2 
	`
	// если формула уже есть в таблице, повторный INSERT
	// упадёт с ошибкой нарушения PRIMARY KEY, поэтому ON CONFLICT

	_, err := a.db.Exec(query, clientID, formula)
	if err != nil {
		logger.ErrLog.Printf("ошибка при SQL запросе INSERT: %v", err)
		return fmt.Errorf("insert free formula: %w", err)
	}
	return nil
}

// возвращаем тип формулы и её значение для расчёта
func (a *App) GetClientFormula(clientID int) (Formula, error) {
	var ft sql.NullString
	err := a.db.QueryRow(
		`SELECT formula_type FROM clients WHERE id = $1`, clientID,
	).Scan(&ft)
	if err != nil {
		return nil, fmt.Errorf("get client: %w", err)
	}
	if !ft.Valid {
		return nil, fmt.Errorf("у клиента не задана формула")
	}

	switch ft.String {
	case "standard":
		var percent float64
		err = a.db.QueryRow(
			`SELECT percent FROM standard_formulas WHERE client_id = $1`, clientID,
		).Scan(&percent)
		if err != nil {
			return nil, fmt.Errorf("get standard formula: %w", err)
		}
		return StandardFormula{Percent: percent}, nil

	case "salary":
		var coefficient float64
		err = a.db.QueryRow(
			`SELECT coefficient FROM salary_formulas WHERE client_id = $1`, clientID,
		).Scan(&coefficient)
		if err != nil {
			return nil, fmt.Errorf("get salary formula: %w", err)
		}
		return SalaryFormula{Coefficient: coefficient}, nil

	case "free":
		var text string
		err = a.db.QueryRow(
			`SELECT formula FROM free_formulas WHERE client_id = $1`, clientID,
		).Scan(&text)
		if err != nil {
			return nil, fmt.Errorf("get free formula: %w", err)
		}
		return FreeFormula{Text: text}, nil

	default:
		return nil, fmt.Errorf("неизвестный тип формулы: %s", ft.String)
	}
}

func (a *App) DELETEClient(clientID int) error {
	query := `
	DELETE FROM clients 
	WHERE id = $1`

	result, err := a.db.Exec(query, clientID)
	if err != nil {
		logger.ErrLog.Printf("ошибка DeleteClient: %v", err)
		return fmt.Errorf("delete client: %w", err)
	}

	// Проверяем, сколько строк было удалено
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("клиент с ID %d не найден", clientID)
	}
	return nil
}

// получение данных из таблицы clients
func (a *App) GetClients() ([]Clients, error) {

	query := `
	SELECT id, name, formula_type
	FROM clients
	ORDER BY id
	` //order by - упорядочивание

	//Выбираем QueryRow если возврат одной строки, Query если 0..N строк
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query clients: %w", err)
	}

	//rows держит открытый ресурс (сетевое соединение / курсор на сервере).
	//Если не закрыть, ресурсы утекут и дальнейшие запросы могут упасть
	defer rows.Close()

	//хранение массива КЛИЕНТОВ (структура, НЕ массив string)
	var clients []Clients

	//медот двигает курсор поСТРОЧНО - return `true`, если строка есть, иначе false
	for rows.Next() { // 1. аналог циклов while, двигаемся построчно

		var c Clients // 2. создаём переменную для каждой строки

		err := rows.Scan( // 3. Чтение данных
			&c.ID,
			&c.Name,
			&c.FormulaType,
		)

		if err != nil {
			logger.ErrLog.Printf("ошибка при чтении строки клиента: %v", err)
			return nil, fmt.Errorf("scan client: %w", err)
		}

		clients = append(clients, c) //5. добавляем в слайс
	}

	return clients, nil
}
