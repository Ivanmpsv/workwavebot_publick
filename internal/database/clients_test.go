package database

import (
	"database/sql"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// Запуск: go test ./internal/database/... -run Client -v
// флаг -run фильтрует по имени теста регуляркой, так удобно гонять только тесты клиентов, не трогая admins_test.go

// Хелпер newMockApp и TestMain — общие для всего пакета database,
// см. admins_test.go, где они определены и подробно прокомментированы.

func TestAddClient(t *testing.T) {
	app, mock := newMockApp(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO clients")).
		WithArgs("альфа").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := app.AddClient("альфа"); err != nil {
		t.Fatalf("AddClient() вернул ошибку: %v", err)
	}
}

func TestAddClient_DBError(t *testing.T) {
	app, mock := newMockApp(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO clients")).
		WithArgs("альфа").
		WillReturnError(errDBFailure)

	if err := app.AddClient("альфа"); err == nil {
		t.Fatal("AddClient() не вернул ошибку при сбое БД")
	}
}

// TestChangeFormulaType_ValidatesInputBeforeTouchingDB — важный момент:
// ChangeFormulaType сначала проверяет formulaType через switch/default
// и только потом идёт в БД. Значит для невалидного типа mock НЕ должен
// получить никакого запроса вообще — это и проверяем, не регистрируя
// ни одного mock.ExpectExec.
func TestChangeFormulaType_ValidatesInputBeforeTouchingDB(t *testing.T) {
	app, mock := newMockApp(t)

	err := app.ChangeFormulaType(1, "неизвестный_тип")
	if err == nil {
		t.Fatal("ChangeFormulaType() с невалидным типом должен вернуть ошибку")
	}
	if unmetErr := mock.ExpectationsWereMet(); unmetErr != nil {
		t.Errorf("ChangeFormulaType не должен был обращаться к БД на невалидном вводе: %v", unmetErr)
	}
}

// TestChangeFormulaType_ValidTypes — table-driven: три допустимых значения
// formulaType ("standard", "salary", "free") должны отработать одинаково,
// меняется только само значение типа.
func TestChangeFormulaType_ValidTypes(t *testing.T) {
	validTypes := []string{"standard", "salary", "free"}

	for _, formulaType := range validTypes {
		t.Run(formulaType, func(t *testing.T) {
			app, mock := newMockApp(t)

			mock.ExpectExec(regexp.QuoteMeta("UPDATE clients")).
				WithArgs(1, formulaType).
				WillReturnResult(sqlmock.NewResult(0, 1))

			if err := app.ChangeFormulaType(1, formulaType); err != nil {
				t.Errorf("ChangeFormulaType(1, %q) вернул ошибку: %v", formulaType, err)
			}
		})
	}
}

// TestSetFormula_AllThreeVariants объединяет в одну таблицу тесты трёх
// очень похожих по форме методов: SetStandardFormula, SetSalaryFormula,
// SetFreeFormula. У всех троих одна и та же схема поведения
// (INSERT ... ON CONFLICT DO UPDATE), различаются только имя таблицы/
// колонки и сам вызов — поэтому вместо трёх копипаст-тестов один общий
// прогон с полем "call", которое дёргает нужный метод App.
func TestSetFormula_AllThreeVariants(t *testing.T) {
	cases := []struct {
		name      string
		queryLike string // характерный кусок SQL, который ожидаем увидеть
		call      func(app *App) error
	}{
		{
			name:      "SetStandardFormula",
			queryLike: "INSERT INTO standard_formulas",
			call: func(app *App) error {
				return app.SetStandardFormula(1, 0.15)
			},
		},
		{
			name:      "SetSalaryFormula",
			queryLike: "INSERT INTO salary_formulas",
			call: func(app *App) error {
				return app.SetSalaryFormula(1, 1.0)
			},
		},
		{
			name:      "SetFreeFormula",
			queryLike: "INSERT INTO free_formulas",
			call: func(app *App) error {
				return app.SetFreeFormula(1, "salary * 0.2")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, mock := newMockApp(t)

			mock.ExpectExec(regexp.QuoteMeta(tc.queryLike)).
				WillReturnResult(sqlmock.NewResult(0, 1))

			if err := tc.call(app); err != nil {
				t.Errorf("%s вернул ошибку: %v", tc.name, err)
			}
		})
	}
}

// TestGetClientFormula проверяет самую нетривиальную функцию файла:
// она сначала читает formula_type клиента, а затем — в зависимости от
// типа — идёт за значением в ОДНУ из трёх разных таблиц и возвращает
// соответствующую реализацию интерфейса Formula.
//
// Три кейса standard/salary/free имеют разную форму (разные таблицы,
// разные поля, разный конкретный тип на выходе), поэтому мы не сводим их
// в одну таблицу, а описываем явно — это тот случай, где "давайте сделаем
// таблицу ради таблицы" ухудшил бы читаемость.
func TestGetClientFormula_Standard(t *testing.T) {
	app, mock := newMockApp(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT formula_type FROM clients")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"formula_type"}).AddRow("standard"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT percent FROM standard_formulas")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"percent"}).AddRow(0.15))

	formula, err := app.GetClientFormula(1)
	if err != nil {
		t.Fatalf("GetClientFormula() вернул ошибку: %v", err)
	}

	got, ok := formula.(StandardFormula)
	if !ok {
		t.Fatalf("GetClientFormula() вернул %T, want StandardFormula", formula)
	}
	if got.Percent != 0.15 {
		t.Errorf("StandardFormula.Percent = %v, want 0.15", got.Percent)
	}
}

func TestGetClientFormula_Salary(t *testing.T) {
	app, mock := newMockApp(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT formula_type FROM clients")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"formula_type"}).AddRow("salary"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT coefficient FROM salary_formulas")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"coefficient"}).AddRow(1.0))

	formula, err := app.GetClientFormula(1)
	if err != nil {
		t.Fatalf("GetClientFormula() вернул ошибку: %v", err)
	}

	got, ok := formula.(SalaryFormula)
	if !ok {
		t.Fatalf("GetClientFormula() вернул %T, want SalaryFormula", formula)
	}
	if got.Coefficient != 1.0 {
		t.Errorf("SalaryFormula.Coefficient = %v, want 1.0", got.Coefficient)
	}
}

func TestGetClientFormula_Free(t *testing.T) {
	app, mock := newMockApp(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT formula_type FROM clients")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"formula_type"}).AddRow("free"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT formula FROM free_formulas")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"formula"}).AddRow("salary * 0.2"))

	formula, err := app.GetClientFormula(1)
	if err != nil {
		t.Fatalf("GetClientFormula() вернул ошибку: %v", err)
	}

	got, ok := formula.(FreeFormula)
	if !ok {
		t.Fatalf("GetClientFormula() вернул %T, want FreeFormula", formula)
	}
	if got.Text != "salary * 0.2" {
		t.Errorf("FreeFormula.Text = %q, want %q", got.Text, "salary * 0.2")
	}
}

// TestGetClientFormula_NoFormulaYet — у клиента ещё не выбран тип формулы,
// formula_type в БД это SQL NULL. sql.NullString.Valid == false в этом
// случае, и функция обязана вернуть понятную ошибку, а не паниковать
// на пустой строке.
func TestGetClientFormula_NoFormulaYet(t *testing.T) {
	app, mock := newMockApp(t)

	nullFormulaType := sqlmock.NewRows([]string{"formula_type"}).AddRow(nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT formula_type FROM clients")).
		WithArgs(1).
		WillReturnRows(nullFormulaType)

	_, err := app.GetClientFormula(1)
	if err == nil {
		t.Fatal("GetClientFormula() должен вернуть ошибку, если у клиента ещё нет формулы")
	}
}

// TestGetClientFormula_ClientNotFound — клиента с таким ID вообще нет
// в таблице clients: QueryRow.Scan вернёт sql.ErrNoRows.
func TestGetClientFormula_ClientNotFound(t *testing.T) {
	app, mock := newMockApp(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT formula_type FROM clients")).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	_, err := app.GetClientFormula(999)
	if err == nil {
		t.Fatal("GetClientFormula() должен вернуть ошибку для несуществующего клиента")
	}
}

func TestDELETEClient(t *testing.T) {
	// та же форма кейсов, что и в TestDeleteAdmin из admins_test.go:
	// успех / ошибка Exec / 0 затронутых строк.
	cases := []struct {
		name         string
		execErr      error
		rowsAffected int64
		wantErr      bool
	}{
		{name: "клиент успешно удалён", rowsAffected: 1, wantErr: false},
		{name: "ошибка БД при удалении", execErr: errDBFailure, wantErr: true},
		{name: "клиента с таким ID нет — 0 затронутых строк", rowsAffected: 0, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, mock := newMockApp(t)

			expect := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM clients")).WithArgs(1)

			if tc.execErr != nil {
				expect.WillReturnError(tc.execErr)
			} else {
				expect.WillReturnResult(sqlmock.NewResult(0, tc.rowsAffected))
			}

			err := app.DELETEClient(1)

			if (err != nil) != tc.wantErr {
				t.Fatalf("DELETEClient() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestGetClients(t *testing.T) {
	app, mock := newMockApp(t)

	// formula_type у "бета" ещё NULL (клиент только создан, формула не задана) —
	// это нормальный сценарий, который обязан отрабатывать sql.NullString,
	// а не падать.
	rows := sqlmock.NewRows([]string{"id", "name", "formula_type"}).
		AddRow(1, "альфа", "standard").
		AddRow(2, "бета", nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, formula_type")).WillReturnRows(rows)

	got, err := app.GetClients()
	if err != nil {
		t.Fatalf("GetClients() вернул ошибку: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("GetClients() вернул %d записей, want 2", len(got))
	}

	if got[0].Name != "альфа" || !got[0].FormulaType.Valid || got[0].FormulaType.String != "standard" {
		t.Errorf("GetClients()[0] = %+v, want Name=альфа, FormulaType=standard", got[0])
	}
	if got[1].Name != "бета" || got[1].FormulaType.Valid {
		t.Errorf("GetClients()[1] = %+v, want Name=бета, FormulaType=NULL (Valid=false)", got[1])
	}
}

func TestGetClients_QueryError(t *testing.T) {
	app, mock := newMockApp(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, formula_type")).WillReturnError(errDBFailure)

	_, err := app.GetClients()
	if err == nil {
		t.Fatal("GetClients() не вернул ошибку при сбое запроса")
	}
}
