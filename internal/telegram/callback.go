package telegram // callback.go Callback – константы для callback-ов,
// которые используются для боработки кнопок

type Callback string

const (
	CallBackRecruierMenu Callback = "recruier_menu"
	//кнопки для рекрутера
	CallbackBonus      Callback = "bonus"
	CallbackClients    Callback = "clients"
	CallbackUsefulness Callback = "Usefulness"

	//админские кнопки
	CallbackAdminMenu          Callback = "admin_menu"
	CallbackClientsControlMenu Callback = "clients_control"
	CallbackAdminsControlMenu  Callback = "admins_control"
	CallbackAllAdmins          Callback = "all_admins"
	CallbackAddAdmin           Callback = "add_admin"
	CallbackDeleteAdmin        Callback = "delete_admin"

	CallbackAllClients   Callback = "all_clients"
	CallbackAddClient    Callback = "add_client"
	CallbackUpdateClient Callback = "update_client"
	CallbackDeleteClient Callback = "delete_client"

	//формулы (типы формул)
	CallbackFormulaStandard Callback = "formula:standard"
	CallbackFormulaSalary   Callback = "formula:salary"
	CallbackFormulaFree     Callback = "formula:free"

	CallbackBack Callback = "back"
)
