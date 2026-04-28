package calculator

import (
	"fmt"
	"workwavebot/internal/utils"
)

func StandartFormula(salary, clientPercent float64) string {
	result20 := salary * 12 * clientPercent * 0.2
	result30 := salary * 12 * clientPercent * 0.3

	return fmt.Sprintf("20%% от чека: %s * 12 * %.2f * 20%% = %s \n\n"+
		"30%% от чека: %s * 12 * %.2f * 30%% = %s",
		utils.FormatBonus(salary), clientPercent, utils.FormatBonus(result20),
		utils.FormatBonus(salary), clientPercent, utils.FormatBonus(result30))

}

func SalaryFormula(salary, coefficient float64) string {
	result20 := salary * coefficient * 0.2
	result30 := salary * coefficient * 0.3

	return fmt.Sprintf("20%% от чека: %s * %.2f * 20%% = %s \n\n"+
		"30%% от чека: %s * %.2f * 30%% = %s",
		utils.FormatBonus(salary), coefficient, utils.FormatBonus(result20),
		utils.FormatBonus(salary), coefficient, utils.FormatBonus(result30))
}

func FreeFormula(a string) string {
	// TODO: создать формулу с подключением API ИИ
	return "формула пока в разработке"
}
