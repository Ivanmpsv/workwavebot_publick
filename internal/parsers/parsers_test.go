package parsers

import "testing"

// Запуск: go test ./internal/parsers/... -v
//
// Пакет parsers специально не знает про БД и Telegram (см. комментарий
// в parsers.go про единственную ответственность) — это значит, что все три
// функции тут чистые: string -> (значение, error), без побочных эффектов.
// Именно такие функции идеальны для table-driven тестов: не нужен ни мок,
// ни настоящая БД, ни сеть — только вход и ожидаемый выход.

func TestParseFloat(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    float64
		wantErr bool
	}{
		{name: "обычное дробное число с точкой", in: "0.15", want: 0.15},
		{name: "целое число без дробной части", in: "150000", want: 150000},
		{name: "запятая вместо точки (частая опечатка при вводе с телефона)", in: "0,15", want: 0.15},
		{name: "пробелы вокруг числа", in: "  150000  ", want: 150000},
		{name: "пробел внутри числа (разделитель разрядов) тоже удаляется", in: "150 000", want: 150000},
		{name: "отрицательное число - валидное с точки зрения парсера", in: "-10", want: -10},
		{name: "пустая строка — ошибка", in: "", wantErr: true},
		{name: "текст вместо числа — ошибка", in: "abc", wantErr: true},
		{name: "две точки — невалидное число, ошибка", in: "0.1.5", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseFloat(tc.in)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseFloat(%q) не вернул ошибку, хотя ожидалась", tc.in)
				}
				return // при ожидаемой ошибке значение got не проверяем
			}

			if err != nil {
				t.Fatalf("ParseFloat(%q) вернул неожиданную ошибку: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("ParseFloat(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseAdminInput(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		wantID   int64
		wantName string
		wantErr  bool
	}{
		{
			name:     "стандартный формат из подсказки бота: id + имя",
			in:       "123456789 Иван",
			wantID:   123456789,
			wantName: "Иван",
		},
		{
			name:     "имя из нескольких слов не обрезается (SplitN с лимитом 2)",
			in:       "123456789 Иван Петров",
			wantID:   123456789,
			wantName: "Иван Петров",
		},
		{
			name:     "лишние пробелы по краям строки удаляются",
			in:       "  123456789 Иван  ",
			wantID:   123456789,
			wantName: "Иван",
		},
		{
			name:    "без имени вообще — ошибка",
			in:      "123456789",
			wantErr: true,
		},
		{
			name:    "id не число — ошибка",
			in:      "не_число Иван",
			wantErr: true,
		},
		{
			name:    "пустая строка — ошибка",
			in:      "",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, name, err := ParseAdminInput(tc.in)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseAdminInput(%q) не вернул ошибку, хотя ожидалась", tc.in)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseAdminInput(%q) вернул неожиданную ошибку: %v", tc.in, err)
			}
			if id != tc.wantID || name != tc.wantName {
				t.Errorf("ParseAdminInput(%q) = (%d, %q), want (%d, %q)",
					tc.in, id, name, tc.wantID, tc.wantName)
			}
		})
	}
}

func TestParseInt64(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    int64
		wantErr bool
	}{
		{name: "обычный telegram id", in: "123456789", want: 123456789},
		{name: "пробелы по краям удаляются", in: "  123456789  ", want: 123456789},
		{name: "ноль — валидный int64, не ошибка", in: "0", want: 0},
		{name: "отрицательное число - тоже валидный int64", in: "-5", want: -5},
		{name: "нечисловая строка — ошибка", in: "Иван", wantErr: true},
		{name: "пустая строка — ошибка", in: "", wantErr: true},
		{name: "число с пробелом внутри — ошибка (это не то же самое, что trim по краям)", in: "123 456", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseInt64(tc.in)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseInt64(%q) не вернул ошибку, хотя ожидалась", tc.in)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseInt64(%q) вернул неожиданную ошибку: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("ParseInt64(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}
