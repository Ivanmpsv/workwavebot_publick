package calculator

import (
	"strings"
	"testing"
)

// Запуск: go test ./internal/calculator/... -v

/*
StandartFormula и SalaryFormula возвращают результат в читаемом человеку виде
(что рекрутер увидит в тг).
Поэтому тест не сверяет число напрямую, а проверяет, что итоговая строка
содержит правильно посчитанные и отформатированные суммы 20% и 30% от чека.
Мы ищем подстроки через strings.Contains вместо жёсткого сравнения всей
строки — так тест не переписывать при любой косметической правке текста
сообщения (например, поменяли "от чека" на "от суммы чека").
*/
func TestStandartFormula(t *testing.T) {
	// table-driven: один и тот же вызов StandartFormula(salary, percent),
	// меняются только входные данные и то, что должно попасть в вывод.
	cases := []struct {
		name          string
		salary        float64
		clientPercent float64
		wantContains  []string // подстроки, которые обязаны быть в результате
	}{
		{
			name:          "стандартный кейс: 150000 гросс, 15% годовых от клиента",
			salary:        150000,
			clientPercent: 0.15,
			// 150000 * 12 * 0.15 * 0.2 = 54000; * 0.3 = 81000
			wantContains: []string{"54 000.00", "81 000.00"},
		},
		{
			name:          "нулевая зарплата — оба варианта бонуса нулевые",
			salary:        0,
			clientPercent: 0.15,
			wantContains: []string{
				"20% от чека",
				"30% от чека",
			},
		},
		{
			name:          "маленький процент клиента (5%)",
			salary:        100000,
			clientPercent: 0.05,
			// 100000 * 12 * 0.05 * 0.2 = 12000; * 0.3 = 18000
			wantContains: []string{"12 000.00", "18 000.00"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StandartFormula(tc.salary, tc.clientPercent)
			for _, want := range tc.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("StandartFormula(%v, %v) = %q, want it to contain %q",
						tc.salary, tc.clientPercent, got, want)
				}
			}
		})
	}
}

func TestSalaryFormula(t *testing.T) {
	cases := []struct {
		name         string
		salary       float64
		coefficient  float64
		wantContains []string
	}{
		{
			name:        "коэффициент 1 (один оклад гросс)",
			salary:      150000,
			coefficient: 1,
			// 150000 * 1 * 0.2 = 30000; * 0.3 = 45000
			wantContains: []string{"30 000.00", "45 000.00"},
		},
		{
			name:        "коэффициент 0.5 (половина оклада гросс)",
			salary:      150000,
			coefficient: 0.5,
			// 150000 * 0.5 * 0.2 = 15000; * 0.3 = 22500
			wantContains: []string{"15 000.00", "22 500.00"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SalaryFormula(tc.salary, tc.coefficient)
			for _, want := range tc.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("SalaryFormula(%v, %v) = %q, want it to contain %q",
						tc.salary, tc.coefficient, got, want)
				}
			}
		})
	}
}

// FreeFormula пока не реализована по-настоящему (в TODO — расчёт через внешний ИИ)
// Единственное, что можно проверить — заглушка с сообщением о том, что формула
// в разработке, и что она не паникует на произвольном вводе.
func TestFreeFormula_ReturnsPlaceholder(t *testing.T) {
	inputs := []string{"", "любой текст формулы", "salary * 12 * 0.2"}

	for _, in := range inputs {
		got := FreeFormula(in)
		want := "формула пока в разработке"
		if got != want {
			t.Errorf("FreeFormula(%q) = %q, want %q", in, got, want)
		}
	}
}
