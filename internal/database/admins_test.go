package database

import (
	"os"
	"regexp"
	"testing"
	"workwavebot/internal/logger"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// Запуск: go test ./internal/database/... -v

/*
 Как это работает без настоящего PostgreSQL:
 sqlmock.New() возвращает пару (*sql.DB, sqlmock.Sqlmock). Первое — обычный
 *sql.DB, который App использует как ни в чём не бывало (через
 database/sql/driver он просто не ходит в реальную сеть). Второе — mock,
 на котором мы ЗАРАНЕЕ описываем: "жду вот такой SQL-запрос с вот такими
 аргументами, и на него нужно вернуть вот такой результат/ошибку".
 Если App отправит что-то другое или не отправит запрос вовсе — mock.
ExpectationsWereMet() в конце это заметит.

 Мы не сравниваем SQL-запросы посимвольно (это сделало бы тесты хрупкими
 к любой перестановке пробелов/переносов строк в исходнике) — вместо
 этого matching идёт по регулярке через regexp.QuoteMeta на характерный
 кусок запроса (например "INSERT INTO admins").

 TestMain — точка входа для тестов всего пакета database.
 Нужна по одной причине: многие функции пишут в logger.ErrLog даже
 на успешных сценариях (например, при заметках "не найдено" под капотом),
 а logger.ErrLog — это package-level *log.Logger, который остаётся nil,
 пока явно не вызван logger.Init(). Без этой инициализации первый же вызов
 logger.ErrLog.Printf(...) в тестируемом коде паникует с nil pointer
 dereference — причём не в тесте, а глубоко внутри тестируемой функции,
 что довольно неочевидно при первом знакомстве с проектом.
*/

func TestMain(m *testing.M) {
	if err := logger.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// newMockApp — небольшой тестовый хелпер: поднимает sqlmock и оборачивает
// его в *App через NewApp (см. app.go). t.Cleanup гарантирует закрытие
// соединения после теста, даже если тест упал с t.Fatal на середине.
func newMockApp(t *testing.T) (*App, sqlmock.Sqlmock) {
	t.Helper() // при падении покажет строку вызова newMockApp, а не эту

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return NewApp(db), mock
}

func TestAddAdmin(t *testing.T) {
	// table-driven по сценариям "успех/ошибка БД" — форма одна и та же:
	// готовим ожидание в mock, вызываем AddAdmin, проверяем err == / != nil.
	cases := []struct {
		name    string
		mockErr error // если не nil — mock вернёт эту ошибку на Exec
		wantErr bool
	}{
		{name: "успешное добавление админа", mockErr: nil, wantErr: false},
		{name: "БД вернула ошибку (например, дублирующийся id)", mockErr: errDBFailure, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, mock := newMockApp(t)

			expect := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO admins")).
				WithArgs(int64(300), "хэрингтон")

			if tc.mockErr != nil {
				expect.WillReturnError(tc.mockErr)
			} else {
				expect.WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err := app.AddAdmin(300, "хэрингтон")

			if (err != nil) != tc.wantErr {
				t.Fatalf("AddAdmin() error = %v, wantErr %v", err, tc.wantErr)
			}
			if unmetErr := mock.ExpectationsWereMet(); unmetErr != nil {
				t.Errorf("не все ожидаемые SQL-запросы были выполнены: %v", unmetErr)
			}
		})
	}
}

func TestDeleteAdmin(t *testing.T) {
	// Тут форма кейсов уже другая (нужно эмулировать RowsAffected), поэтому
	// таблица шире: не только успех/ошибка Exec, но и "Exec прошёл успешно,
	// но ни одна строка не удалилась" — с точки зрения БД это не ошибка,
	// а с точки зрения бизнес-логики DeleteAdmin — это ошибка "не найден".
	cases := []struct {
		name         string
		execErr      error
		rowsAffected int64
		wantErr      bool
	}{
		{name: "админ успешно удалён", rowsAffected: 1, wantErr: false},
		{name: "запрос на удаление упал с ошибкой БД", execErr: errDBFailure, wantErr: true},
		{name: "запрос прошёл, но админ с таким id не найден (0 строк)", rowsAffected: 0, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, mock := newMockApp(t)

			expect := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM admins")).
				WithArgs(int64(300))

			if tc.execErr != nil {
				expect.WillReturnError(tc.execErr)
			} else {
				expect.WillReturnResult(sqlmock.NewResult(0, tc.rowsAffected))
			}

			err := app.DeleteAdmin(300)

			if (err != nil) != tc.wantErr {
				t.Fatalf("DeleteAdmin() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestGetAdmins(t *testing.T) {
	app, mock := newMockApp(t)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(int64(1), "иванов").
		AddRow(int64(2), "петров")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name FROM admins")).WillReturnRows(rows)

	got, err := app.GetAdmins()
	if err != nil {
		t.Fatalf("GetAdmins() вернул ошибку: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("GetAdmins() вернул %d записей, want 2", len(got))
	}

	want := []Admins{
		{ID: 1, Name: "иванов"},
		{ID: 2, Name: "петров"},
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GetAdmins()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestGetAdmins_QueryError(t *testing.T) {
	app, mock := newMockApp(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name FROM admins")).WillReturnError(errDBFailure)

	_, err := app.GetAdmins()
	if err == nil {
		t.Fatal("GetAdmins() не вернул ошибку при сбое запроса")
	}
}

// TestCheckAdmin — ещё один хороший пример table-driven: единственное,
// что меняется между кейсами — что возвращает БД в столбце "exists"
// (true/false), а форма проверки одна и та же.
func TestCheckAdmin(t *testing.T) {
	cases := []struct {
		name       string
		existsInDB bool
		want       bool
	}{
		{"пользователь есть в таблице admins", true, true},
		{"пользователя нет в таблице admins", false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, mock := newMockApp(t)

			rows := sqlmock.NewRows([]string{"exists"}).AddRow(tc.existsInDB)
			mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
				WithArgs(int64(42)).
				WillReturnRows(rows)

			if got := app.CheckAdmin(42); got != tc.want {
				t.Errorf("CheckAdmin(42) = %v, want %v", got, tc.want)
			}
		})
	}
}

// CheckAdmin намеренно "проглатывает" любую ошибку БД и возвращает false —
// с точки зрения безопасности это правильное поведение по умолчанию
// (сомневаешься — не давай прав админа), и этот тест фиксирует именно его.
func TestCheckAdmin_DBErrorMeansNotAdmin(t *testing.T) {
	app, mock := newMockApp(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WithArgs(int64(42)).
		WillReturnError(errDBFailure)

	if got := app.CheckAdmin(42); got != false {
		t.Errorf("CheckAdmin(42) при ошибке БД = %v, want false (fail-safe поведение)", got)
	}
}

// errDBFailure — общая "тестовая" ошибка, которой мы имитируем сбой
// на стороне БД (обрыв соединения, дедлок и т.п.) — сути SQL-ошибки для
// наших тестов не важна, важно только что Exec/Query вернули err != nil.
var errDBFailure = fakeErr("simulated db failure")

type fakeErr string

func (e fakeErr) Error() string { return string(e) }
