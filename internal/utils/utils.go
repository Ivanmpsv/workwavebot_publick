package utils // всякие полезности

import "fmt"

// читаемый вывод бонусов: 1050000.50 -> "1 050 000.50"
func FormatBonus(amount float64) string {
	// разбиваем на целую и дробную части
	intPart := int64(amount)
	fracPart := amount - float64(intPart)

	// форматируем дробную часть — всегда два знака
	frac := fmt.Sprintf("%.2f", fracPart)[1:] // берём ".50" из "0.50"

	// целую часть превращаем в строку и добавляем пробелы каждые 3 символа справа
	s := fmt.Sprintf("%d", intPart) // "1050000"

	// вставляем пробелы справа налево через len()
	result := ""
	for i, ch := range s {
		// считаем сколько символов осталось до конца
		remaining := len(s) - i
		// если remaining делится на 3 и мы не в начале — ставим пробел
		if remaining%3 == 0 && i != 0 {
			result += " "
		}
		result += string(ch)
	}

	return result + frac // "1 050 000" + ".50"
}
