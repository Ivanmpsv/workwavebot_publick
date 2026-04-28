package parsers // набор функций для разбора пользовательского ввода
//  Помни: парсер не должен знать о БД – это нарушение принципа единственной ответственности

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseFloat — парсит строку в float64, например "0.15" или "15"
func ParseFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, " ", "")  // удаляет все пробелы/табуляцию
	s = strings.ReplaceAll(s, ",", ".") // на случай если напишут "0,15"

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("ожидается число, получено: %q", s)
	}
	return v, nil
}

// ParseAdminInput — парсит "1234 Стас" → (id, name)
func ParseAdminInput(s string) (int64, string, error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, " ", 2)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("ожидается 'id имя', получено: %q", s)
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("некорректный ID: %w", err)
	}

	return id, strings.TrimSpace(parts[1]), nil
}

// string -> int64
func ParseInt64(s string) (int64, error) {
	s = strings.TrimSpace(s)
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("ожидается числовой ID, получено: %q", s)
	}
	return id, nil
}
