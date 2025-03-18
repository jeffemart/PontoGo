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
/editar <ID> <quantidade> <data> <observação> <retirada> - Edita o banco de horas de um colaborador

Exemplo de edição:
/editar 59 2.5 2023-05-15 "Horas extras" false

Parâmetros:
- ID: ID do registro no banco de horas
- quantidade: Valor em horas (use ponto para decimais)
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
	args := strings.SplitN(message.Text, " ", 6)
	if len(args) < 6 {
		helpMsg := tgbotapi.NewMessage(message.Chat.ID,
			"Formato incorreto. Use:\n/editar <ID> <quantidade> <data> <observação> <retirada>\n\nExemplo:\n/editar 59 2.5 2023-05-15 \"Horas extras\" false")
		b.api.Send(helpMsg)
		return
	}

	// Extrai os argumentos
	entryID := args[1]

	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A quantidade deve ser um número válido (use ponto para decimais).")
		b.api.Send(errorMsg)
		return
	}

	date := args[3]
	_, err = time.Parse("2006-01-02", date)
	if err != nil {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A data deve estar no formato YYYY-MM-DD.")
		b.api.Send(errorMsg)
		return
	}

	// Extrai a observação (pode conter espaços)
	observationParts := strings.SplitN(args[4], "\"", 3)
	var observation string
	if len(observationParts) >= 3 {
		observation = observationParts[1]
	} else {
		observation = args[4]
	}

	// Extrai o último argumento (withdraw)
	var withdraw bool
	withdrawStr := strings.ToLower(args[5])
	if withdrawStr == "true" {
		withdraw = true
	} else if withdrawStr == "false" {
		withdraw = false
	} else {
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: O parâmetro 'retirada' deve ser 'true' ou 'false'.")
		b.api.Send(errorMsg)
		return
	}

	// Cria a entrada para atualização
	entry := models.TimeBalanceEntry{
		Amount:      amount,
		Date:        date,
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

	// Envia mensagem de sucesso
	successMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Banco de horas atualizado com sucesso!\n\nID: %s\nQuantidade: %.2f horas\nData: %s\nObservação: %s\nRetirada: %t",
		entryID, amount, date, observation, withdraw))
	b.api.Send(successMsg)
}
