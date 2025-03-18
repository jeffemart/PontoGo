package models

// Estrutura para armazenar as variáveis de ambiente
type Config struct {
	PontoMaisToken   string
	PontoMaisBaseURL string
	TelegramBotToken string
	TelegramHosts    []int64
	Debug            bool
}

// Estrutura para armazenar os dados do colaborador
type Employee struct {
	ID                int    `json:"id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
	PIN               string `json:"pin"`
	IsCLT             bool   `json:"is_clt"`
	CPF               string `json:"cpf"`
	NIS               string `json:"nis"`
	RegistrationNumber string `json:"registration_number"`
	TimeCardSource    string `json:"time_card_source"`
	HasTimeCards      bool   `json:"has_time_cards"`
	UseQRCode         bool   `json:"use_qrcode"`
	EnableGeolocation bool   `json:"enable_geolocation"`
	WorkHours         string `json:"work_hours"`
	CostCenter        string `json:"cost_center"`
	User              string `json:"user"`
	EnableOfflineTimeCards bool `json:"enable_offline_time_cards"`
	Login             string `json:"login"`
}

// Struct para mapear a resposta da API
type EmployeesResponse struct {
	Employees []Employee `json:"employees"`
}

// TimeBalanceEntry representa os dados para atualização do banco de horas
type TimeBalanceEntry struct {
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Observation string  `json:"observation"`
	Withdraw    bool    `json:"withdraw"`
}