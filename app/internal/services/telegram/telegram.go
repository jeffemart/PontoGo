package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jeffemart/PontoGo/app/internal/models"
	services "github.com/jeffemart/PontoGo/app/internal/services/pontomais"
)

// Bot representa a estrutura do bot do Telegram
type Bot struct {
	api    *tgbotapi.BotAPI
	config *models.Config
	hosts  map[int64]bool
}

// NewBot cria uma nova instância do bot do Telegram
func NewBot(cfg *models.Config) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, err
	}

	// Configura os hosts autorizados
	hosts := make(map[int64]bool)
	for _, hostID := range cfg.TelegramHosts {
		hosts[hostID] = true
	}

	return &Bot{
		api:    bot,
		config: cfg,
		hosts:  hosts,
	}, nil
}

// Start inicia o bot do Telegram
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Erro ao iniciar o bot: %v", err)
	}

	log.Printf("Bot iniciado com sucesso: @%s", b.api.Self.UserName)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Verifica se o usuário está autorizado
		if !b.hosts[update.Message.Chat.ID] {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Você não está autorizado a usar este bot.")
			b.api.Send(msg)
			continue
		}

		// Processa os comandos
		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
		}
	}
}

// handleCommand processa os comandos recebidos pelo bot
func (b *Bot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		b.handleStart(message)
	case "help":
		b.handleHelp(message)
	case "listar":
		b.handleListEmployees(message)
	case "editar":
		b.handleEditTimeBalance(message)
	case "criar":
		b.handleCreateTimeBalance(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Comando desconhecido. Use /help para ver os comandos disponíveis.")
		b.api.Send(msg)
	}
}

// handleStart envia uma mensagem de boas-vindas
func (b *Bot) handleStart(message *tgbotapi.Message) {
	welcomeText := fmt.Sprintf("Olá, %s! Bem-vindo ao PontoGo Bot.\n\nUse /help para ver os comandos disponíveis.", message.From.FirstName)
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	b.api.Send(msg)
}

// handleHelp envia a lista de comandos disponíveis
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	helpText := `Comandos disponíveis:

/start - Inicia o bot
/help - Mostra esta mensagem de ajuda
/listar - Lista todos os colaboradores ativos
/editar <ID> <quantidade_segundos> <data> <observação> <retirada> - Edita o banco de horas de um colaborador
/criar <ID_funcionário> <quantidade_segundos> <data> <observação> <retirada> - Cria um novo lançamento no banco de horas

Exemplo de edição:
/editar 59 9000.0 2023-05-15 "2.5 horas extras" false

Exemplo de criação:
/criar 1487972 60.0 2023-05-15 "1 minuto de trabalho" false

Parâmetros:
- ID: ID do registro no banco de horas (para edição)
- ID_funcionário: ID do funcionário (para criação)
- quantidade_segundos: Valor em segundos (use ponto para decimais)
- data: Data no formato YYYY-MM-DD
- observação: Texto entre aspas com o motivo
- retirada: true para retirada, false para adição`

	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
	b.api.Send(msg)
}

// handleListEmployees lista todos os colaboradores ativos
func (b *Bot) handleListEmployees(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Buscando colaboradores...")
	b.api.Send(msg)

	employees, err := services.GetEmployees(b.config)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Erro ao buscar colaboradores: %v", err))
		b.api.Send(errorMsg)
		return
	}

	if len(employees) == 0 {
		noEmployeesMsg := tgbotapi.NewMessage(message.Chat.ID, "Nenhum colaborador encontrado.")
		b.api.Send(noEmployeesMsg)
		return
	}

	var response string
	if len(employees) == 0 {
		response = "Nenhum funcionário encontrado."
	} else {
		// Retorna apenas a quantidade de funcionários
		response = fmt.Sprintf("Total de funcionários: %d", len(employees))
	}

	resultMsg := tgbotapi.NewMessage(message.Chat.ID, response)
	b.api.Send(resultMsg)
}

// handleEditTimeBalance edita o banco de horas de um colaborador
func (b *Bot) handleEditTimeBalance(message *tgbotapi.Message) {
	// Dividimos a mensagem em partes para extrair os argumentos básicos
	parts := strings.SplitN(message.Text, " ", 5)
	if len(parts) < 5 {
		helpMsg := tgbotapi.NewMessage(message.Chat.ID,
			"Formato incorreto. Use:\n/editar <ID> <quantidade_segundos> <data> <observação> <retirada>\n\nExemplo:\n/editar 3833376 60.0 2025-03-18 \"Editando lançamento\" false")
		b.api.Send(helpMsg)
		return
	}

	// Extrai os argumentos básicos
	entryID := parts[1]

	// Converte a quantidade
	amount, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A quantidade deve ser um número válido (use ponto para decimais).")
		b.api.Send(errorMsg)
		return
	}

	// Formata a data para o formato esperado pela API (DD/MM/YYYY)
	inputDate := parts[3]
	parsedDate, err := time.Parse("2006-01-02", inputDate)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A data deve estar no formato YYYY-MM-DD.")
		b.api.Send(errorMsg)
		return
	}
	formattedDate := parsedDate.Format("02/01/2006")

	// Extrai a parte final que contém a observação e o parâmetro de retirada
	lastPart := parts[4]

	// Procura a última ocorrência de aspas para separar a observação do parâmetro de retirada
	lastQuoteIndex := strings.LastIndex(lastPart, "\"")
	if lastQuoteIndex == -1 || lastQuoteIndex == 0 {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A observação deve estar entre aspas duplas.")
		b.api.Send(errorMsg)
		return
	}

	// Encontra a primeira ocorrência de aspas
	firstQuoteIndex := strings.Index(lastPart, "\"")
	if firstQuoteIndex == lastQuoteIndex {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A observação deve estar entre aspas duplas.")
		b.api.Send(errorMsg)
		return
	}

	// Extrai a observação entre aspas
	observation := lastPart[firstQuoteIndex+1 : lastQuoteIndex]

	// Extrai o parâmetro de retirada após a última aspas
	withdrawPart := strings.TrimSpace(lastPart[lastQuoteIndex+1:])

	// Verifica se o parâmetro de retirada é válido
	var withdraw bool
	if withdrawPart == "true" {
		withdraw = true
	} else if withdrawPart == "false" {
		withdraw = false
	} else {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: O parâmetro 'retirada' deve ser 'true' ou 'false'.")
		b.api.Send(errorMsg)
		return
	}

	// Cria a entrada para atualização
	entry := models.TimeBalanceEntry{
		Amount:      amount,
		Date:        formattedDate,
		Observation: observation,
		Withdraw:    withdraw,
	}

	// Envia mensagem de processamento
	processingMsg := tgbotapi.NewMessage(message.Chat.ID, "Processando atualização do banco de horas...")
	b.api.Send(processingMsg)

	// Atualiza o banco de horas
	err = services.UpdateTimeBalanceEntry(b.config, entryID, entry)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Erro ao atualizar o banco de horas: %v", err))
		b.api.Send(errorMsg)
		return
	}

	// Calcula as horas para exibição
	hoursAmount := amount / 3600.0

	// Envia mensagem de sucesso
	successMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Banco de horas atualizado com sucesso!\n\nID: %s\nQuantidade: %.2f segundos (%.2f horas)\nData: %s\nObservação: %s\nRetirada: %t",
		entryID, amount, hoursAmount, inputDate, observation, withdraw))
	b.api.Send(successMsg)
}

// handleCreateTimeBalance cria um novo lançamento no banco de horas de um funcionário
func (b *Bot) handleCreateTimeBalance(message *tgbotapi.Message) {
	// Dividimos a mensagem em partes para extrair os argumentos básicos
	parts := strings.SplitN(message.Text, " ", 5)
	if len(parts) < 5 {
		helpMsg := tgbotapi.NewMessage(message.Chat.ID,
			"Formato incorreto. Use:\n/criar <ID_funcionário> <quantidade_segundos> <data> <observação> <retirada>\n\nExemplo:\n/criar 1487972 3600.0 2023-05-15 \"1 hora de trabalho\" false")
		b.api.Send(helpMsg)
		return
	}

	// Extrai os argumentos básicos
	employeeID := parts[1]

	// Converte a quantidade diretamente em segundos (sem multiplicar por 3600)
	secondsAmount, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A quantidade deve ser um número válido (use ponto para decimais).")
		b.api.Send(errorMsg)
		return
	}

	// Calcula as horas para exibição na mensagem de sucesso
	hoursAmount := secondsAmount / 3600.0

	// Formata a data
	inputDate := parts[3]
	parsedDate, err := time.Parse("2006-01-02", inputDate)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A data deve estar no formato YYYY-MM-DD.")
		b.api.Send(errorMsg)
		return
	}
	formattedDate := parsedDate.Format("02/01/2006")

	// Extrai a parte final que contém a observação e o parâmetro de retirada
	lastPart := parts[4]

	// Procura a última ocorrência de aspas para separar a observação do parâmetro de retirada
	lastQuoteIndex := strings.LastIndex(lastPart, "\"")
	if lastQuoteIndex == -1 || lastQuoteIndex == 0 {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A observação deve estar entre aspas duplas.")
		b.api.Send(errorMsg)
		return
	}

	// Encontra a primeira ocorrência de aspas
	firstQuoteIndex := strings.Index(lastPart, "\"")
	if firstQuoteIndex == lastQuoteIndex {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A observação deve estar entre aspas duplas.")
		b.api.Send(errorMsg)
		return
	}

	// Extrai a observação entre aspas
	observation := lastPart[firstQuoteIndex+1 : lastQuoteIndex]

	// Extrai o parâmetro de retirada após a última aspas
	withdrawPart := strings.TrimSpace(lastPart[lastQuoteIndex+1:])

	// Verifica se o parâmetro de retirada é válido
	var withdraw bool
	if withdrawPart == "true" {
		withdraw = true
	} else if withdrawPart == "false" {
		withdraw = false
	} else {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: O parâmetro 'retirada' deve ser 'true' ou 'false'.")
		b.api.Send(errorMsg)
		return
	}

	// Cria a entrada para o banco de horas
	entry := models.TimeBalanceEntry{
		Amount:      secondsAmount,
		Date:        formattedDate,
		EmployeeID:  employeeID,
		Observation: observation,
		Withdraw:    withdraw,
	}

	// Envia mensagem de processamento
	processingMsg := tgbotapi.NewMessage(message.Chat.ID, "Processando criação do lançamento no banco de horas...")
	b.api.Send(processingMsg)

	// Cria o lançamento no banco de horas
	err = services.CreateTimeBalanceEntry(b.config, entry)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Erro ao criar o lançamento no banco de horas: %v", err))
		b.api.Send(errorMsg)
		return
	}

	// Envia mensagem de sucesso com a conversão para horas para melhor visualização
	successMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Lançamento no banco de horas criado com sucesso!\n\nFuncionário ID: %s\nQuantidade: %.2f segundos (%.2f horas)\nData: %s\nObservação: %s\nRetirada: %t",
		employeeID, secondsAmount, hoursAmount, inputDate, observation, withdraw))
	b.api.Send(successMsg)
}
