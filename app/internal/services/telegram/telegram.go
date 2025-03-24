package telegram

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jeffemart/PontoGo/app/internal/models"
	services "github.com/jeffemart/PontoGo/app/internal/services/pontomais"
	"github.com/jeffemart/PontoGo/app/internal/utils"
	"github.com/xuri/excelize/v2"
)

// Bot representa a estrutura do bot do Telegram
type Bot struct {
	api              *tgbotapi.BotAPI
	config           *models.Config
	hosts            map[int64]bool
	awaitingDocument map[int64]string // Mapa para rastrear usuários aguardando documentos
}

// NewBot cria uma nova instância do bot do Telegram
func NewBot(cfg *models.Config) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		utils.Logger.Printf("Erro ao criar o bot do Telegram: %v", err)
		return nil, err
	}

	// Configura os hosts autorizados
	hosts := make(map[int64]bool)
	for _, hostID := range cfg.TelegramHosts {
		hosts[hostID] = true
	}

	utils.Logger.Printf("Bot do Telegram criado com sucesso: @%s", bot.Self.UserName)
	return &Bot{
		api:              bot,
		config:           cfg,
		hosts:            hosts,
		awaitingDocument: make(map[int64]string),
	}, nil
}

// Start inicia o bot do Telegram
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		utils.Logger.Fatalf("Erro ao iniciar o bot: %v", err)
	}

	utils.Logger.Printf("Bot iniciado com sucesso: @%s", b.api.Self.UserName)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Verifica se o usuário está autorizado
		if !b.hosts[update.Message.Chat.ID] {
			utils.Logger.Printf("Tentativa de acesso não autorizado do chat ID: %d", update.Message.Chat.ID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Você não está autorizado a usar este bot.")
			b.api.Send(msg)
			continue
		}

		// Verifica se o usuário está aguardando um documento
		if command, ok := b.awaitingDocument[update.Message.Chat.ID]; ok {
			if update.Message.Document != nil {
				// Processa o documento recebido
				b.handleDocumentReceived(update.Message, command)
				// Remove o usuário da lista de espera
				delete(b.awaitingDocument, update.Message.Chat.ID)
			} else {
				// Se não for um documento, envia uma mensagem de erro
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Por favor, envie um arquivo Excel (.xlsx).")
				b.api.Send(msg)
			}
			continue
		}

		// Processa os comandos
		if update.Message.IsCommand() {
			utils.Logger.Printf("Comando recebido: %s do chat ID: %d", update.Message.Command(), update.Message.Chat.ID)
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
	case "relatorio":
		b.handleRelatorio(message)
	default:
		utils.Logger.Printf("Comando desconhecido recebido: %s", message.Command())
		msg := tgbotapi.NewMessage(message.Chat.ID, "Comando desconhecido. Use /help para ver os comandos disponíveis.")
		b.api.Send(msg)
	}
}

// handleStart envia uma mensagem de boas-vindas
func (b *Bot) handleStart(message *tgbotapi.Message) {
	utils.Logger.Printf("Comando /start recebido do chat ID: %d", message.Chat.ID)
	welcomeText := fmt.Sprintf("Olá, %s! Bem-vindo ao PontoGo Bot.\n\nUse /help para ver os comandos disponíveis.", message.From.FirstName)
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	b.api.Send(msg)
}

// handleHelp envia a lista de comandos disponíveis
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	utils.Logger.Printf("Comando /help recebido do chat ID: %d", message.Chat.ID)
	helpText := `Comandos disponíveis:

/start - Inicia o bot
/help - Mostra esta mensagem de ajuda
/listar - Lista todos os colaboradores ativos
/editar <ID> <quantidade_segundos> <data> <observação> <retirada> - Edita o banco de horas de um colaborador
/criar <ID_funcionário> <quantidade_segundos> <data> <observação> <retirada> - Cria um novo lançamento no banco de horas
/relatorio - Permite processar múltiplos lançamentos de banco de horas a partir de um arquivo Excel.

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
	utils.Logger.Printf("Comando /listar recebido do chat ID: %d", message.Chat.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Buscando colaboradores...")
	b.api.Send(msg)

	employees, err := services.GetEmployees(b.config)
	if err != nil {
		utils.Logger.Printf("Erro ao buscar colaboradores: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Erro ao buscar colaboradores: %v", err))
		b.api.Send(errorMsg)
		return
	}

	if len(employees) == 0 {
		utils.Logger.Println("Nenhum colaborador encontrado")
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

	utils.Logger.Printf("Encontrados %d colaboradores", len(employees))
	resultMsg := tgbotapi.NewMessage(message.Chat.ID, response)
	b.api.Send(resultMsg)
}

// handleEditTimeBalance edita o banco de horas de um colaborador
func (b *Bot) handleEditTimeBalance(message *tgbotapi.Message) {
	utils.Logger.Printf("Comando /editar recebido do chat ID: %d", message.Chat.ID)
	// Dividimos a mensagem em partes para extrair os argumentos básicos
	parts := strings.SplitN(message.Text, " ", 5)
	if len(parts) < 5 {
		utils.Logger.Printf("Formato incorreto para o comando /editar: %s", message.Text)
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
		utils.Logger.Printf("Erro ao converter quantidade para float: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A quantidade deve ser um número válido (use ponto para decimais).")
		b.api.Send(errorMsg)
		return
	}

	// Formata a data para o formato esperado pela API (DD/MM/YYYY)
	inputDate := parts[3]
	parsedDate, err := time.Parse("2006-01-02", inputDate)
	if err != nil {
		utils.Logger.Printf("Erro ao converter data: %v", err)
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
		utils.Logger.Println("Erro: Observação não está entre aspas duplas")
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A observação deve estar entre aspas duplas.")
		b.api.Send(errorMsg)
		return
	}

	// Encontra a primeira ocorrência de aspas
	firstQuoteIndex := strings.Index(lastPart, "\"")
	if firstQuoteIndex == lastQuoteIndex {
		utils.Logger.Println("Erro: Observação não está entre aspas duplas")
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
		utils.Logger.Printf("Erro: Parâmetro de retirada inválido: %s", withdrawPart)
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
	utils.Logger.Printf("Atualizando banco de horas para o ID: %s", entryID)
	err = services.UpdateTimeBalanceEntry(b.config, entryID, entry)
	if err != nil {
		utils.Logger.Printf("Erro ao atualizar o banco de horas: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Erro ao atualizar o banco de horas: %v", err))
		b.api.Send(errorMsg)
		return
	}

	// Calcula as horas para exibição
	hoursAmount := amount / 3600.0

	// Envia mensagem de sucesso
	utils.Logger.Printf("Banco de horas atualizado com sucesso para o ID: %s", entryID)
	successMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Banco de horas atualizado com sucesso!\n\nID: %s\nQuantidade: %.2f segundos (%.2f horas)\nData: %s\nObservação: %s\nRetirada: %t",
		entryID, amount, hoursAmount, inputDate, observation, withdraw))
	b.api.Send(successMsg)
}

// handleCreateTimeBalance cria um novo lançamento no banco de horas de um funcionário
func (b *Bot) handleCreateTimeBalance(message *tgbotapi.Message) {
	utils.Logger.Printf("Comando /criar recebido do chat ID: %d", message.Chat.ID)
	// Dividimos a mensagem em partes para extrair os argumentos básicos
	parts := strings.SplitN(message.Text, " ", 5)
	if len(parts) < 5 {
		utils.Logger.Printf("Formato incorreto para o comando /criar: %s", message.Text)
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
		utils.Logger.Printf("Erro ao converter quantidade para float: %v", err)
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
		utils.Logger.Printf("Erro ao converter data: %v", err)
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
		utils.Logger.Println("Erro: Observação não está entre aspas duplas")
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro: A observação deve estar entre aspas duplas.")
		b.api.Send(errorMsg)
		return
	}

	// Encontra a primeira ocorrência de aspas
	firstQuoteIndex := strings.Index(lastPart, "\"")
	if firstQuoteIndex == lastQuoteIndex {
		utils.Logger.Println("Erro: Observação não está entre aspas duplas")
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
		utils.Logger.Printf("Erro: Parâmetro de retirada inválido: %s", withdrawPart)
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
	utils.Logger.Printf("Criando lançamento no banco de horas para o funcionário ID: %s", employeeID)
	err = services.CreateTimeBalanceEntry(b.config, entry)
	if err != nil {
		utils.Logger.Printf("Erro ao criar o lançamento no banco de horas: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Erro ao criar o lançamento no banco de horas: %v", err))
		b.api.Send(errorMsg)
		return
	}

	// Envia mensagem de sucesso com a conversão para horas para melhor visualização
	utils.Logger.Printf("Lançamento no banco de horas criado com sucesso para o funcionário ID: %s", employeeID)
	successMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Lançamento no banco de horas criado com sucesso!\n\nFuncionário ID: %s\nQuantidade: %.2f segundos (%.2f horas)\nData: %s\nObservação: %s\nRetirada: %t",
		employeeID, secondsAmount, hoursAmount, inputDate, observation, withdraw))
	b.api.Send(successMsg)
}

// handleRelatorio solicita ao usuário que envie um arquivo Excel
func (b *Bot) handleRelatorio(message *tgbotapi.Message) {
	utils.Logger.Printf("Comando /relatorio recebido do chat ID: %d", message.Chat.ID)

	// Marca o usuário como aguardando um documento
	b.awaitingDocument[message.Chat.ID] = "relatorio"

	// Envia uma mensagem para o usuário solicitando o arquivo
	msg := tgbotapi.NewMessage(message.Chat.ID, "Por favor, envie o arquivo Excel com os dados.")
	b.api.Send(msg)
}

// handleDocumentReceived processa o documento recebido após um comando
func (b *Bot) handleDocumentReceived(message *tgbotapi.Message, command string) {
	utils.Logger.Printf("Documento recebido do chat ID: %d para o comando: %s", message.Chat.ID, command)

	// Obtém o arquivo do Telegram
	fileID := message.Document.FileID
	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		utils.Logger.Printf("Erro ao obter o arquivo: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro ao obter o arquivo.")
		b.api.Send(errorMsg)
		return
	}

	// Baixa o arquivo
	filePath := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.api.Token, file.FilePath)
	resp, err := http.Get(filePath)
	if err != nil {
		utils.Logger.Printf("Erro ao baixar o arquivo: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro ao baixar o arquivo.")
		b.api.Send(errorMsg)
		return
	}
	defer resp.Body.Close()

	// Cria um arquivo temporário em vez de usar um caminho fixo
	tempFile, err := os.CreateTemp("", "excel-*.xlsx")
	if err != nil {
		utils.Logger.Printf("Erro ao criar arquivo temporário: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro ao processar o arquivo.")
		b.api.Send(errorMsg)
		return
	}
	defer os.Remove(tempFile.Name()) // Garante que o arquivo será removido ao final
	defer tempFile.Close()

	// Copia o conteúdo do arquivo baixado para o arquivo temporário
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		utils.Logger.Printf("Erro ao salvar arquivo temporário: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro ao salvar o arquivo.")
		b.api.Send(errorMsg)
		return
	}

	// Fecha o arquivo para garantir que todos os dados foram escritos
	tempFile.Close()

	// Processa o arquivo de acordo com o comando
	if command == "relatorio" {
		b.processRelatorioFile(message, tempFile.Name())
	}
}

// processRelatorioFile processa o arquivo Excel do relatório
func (b *Bot) processRelatorioFile(message *tgbotapi.Message, filePath string) {
	// Lê o arquivo Excel
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		utils.Logger.Printf("Erro ao abrir o arquivo Excel: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro ao abrir o arquivo Excel.")
		b.api.Send(errorMsg)
		return
	}
	defer f.Close()

	// Lê as linhas da primeira planilha
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		utils.Logger.Printf("Erro ao ler as linhas da planilha: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Erro ao ler as linhas da planilha.")
		b.api.Send(errorMsg)
		return
	}

	// Verifica se há linhas suficientes
	if len(rows) < 2 {
		utils.Logger.Println("Arquivo Excel não contém dados suficientes")
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "O arquivo Excel não contém dados suficientes.")
		b.api.Send(errorMsg)
		return
	}

	// Imprime as linhas no console para debug
	utils.Logger.Println("Dados do arquivo Excel:")
	for i, row := range rows {
		utils.Logger.Printf("Linha %d: %v", i, row)
	}

	// Identifica os índices das colunas com base nos cabeçalhos
	headers := rows[0]
	headerMap := make(map[string]int)

	for i, header := range headers {
		headerMap[strings.ToUpper(strings.TrimSpace(header))] = i
	}

	// Verifica se todas as colunas necessárias estão presentes
	requiredHeaders := []string{"ID", "NOME", "DATA", "HORAS", "OBSERVAÇÃO", "DEBITO"}
	missingHeaders := []string{}

	for _, header := range requiredHeaders {
		if _, exists := headerMap[header]; !exists {
			missingHeaders = append(missingHeaders, header)
		}
	}

	if len(missingHeaders) > 0 {
		errorMsg := fmt.Sprintf("Colunas obrigatórias não encontradas: %s", strings.Join(missingHeaders, ", "))
		utils.Logger.Println(errorMsg)
		msg := tgbotapi.NewMessage(message.Chat.ID, errorMsg)
		b.api.Send(msg)
		return
	}

	// Envia mensagem de processamento
	processingMsg := tgbotapi.NewMessage(message.Chat.ID, "Processando lançamentos no banco de horas. Isso pode levar alguns instantes...")
	b.api.Send(processingMsg)

	// Processa cada linha (exceto o cabeçalho)
	successCount := 0
	errorCount := 0
	errorDetails := make([]string, 0)

	for i, row := range rows {
		if i == 0 {
			continue // Pula o cabeçalho
		}

		// Verifica se a linha tem dados suficientes
		if len(row) < len(headers) {
			errorMsg := fmt.Sprintf("Linha %d: Dados insuficientes", i)
			utils.Logger.Println(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			errorCount++
			continue
		}

		// Extrai os dados da linha usando os índices do mapa de cabeçalhos
		employeeID := row[headerMap["ID"]]
		employeeName := row[headerMap["NOME"]]
		dateStr := row[headerMap["DATA"]]
		secondsStr := row[headerMap["HORAS"]] // Agora tratamos como segundos, não horas
		observation := row[headerMap["OBSERVAÇÃO"]]
		debitStr := row[headerMap["DEBITO"]]

		// Converte a data para o formato correto
		date, err := time.Parse("02/01/2006", dateStr)
		if err != nil {
			// Tenta outro formato de data
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				errorMsg := fmt.Sprintf("Linha %d (%s): Formato de data inválido '%s'", i, employeeName, dateStr)
				utils.Logger.Println(errorMsg)
				errorDetails = append(errorDetails, errorMsg)
				errorCount++
				continue
			}
		}
		formattedDate := date.Format("02/01/2006")

		// Converte a string de segundos para float
		seconds, err := strconv.ParseFloat(secondsStr, 64)
		if err != nil {
			errorMsg := fmt.Sprintf("Linha %d (%s): Valor de segundos inválido '%s'", i, employeeName, secondsStr)
			utils.Logger.Println(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			errorCount++
			continue
		}

		// Calcula as horas para exibição (apenas para logs)
		hours := seconds / 3600

		// Determina se é uma retirada
		withdraw := strings.ToLower(debitStr) == "true" || strings.ToLower(debitStr) == "sim" || strings.ToLower(debitStr) == "s"

		// Cria a entrada para o banco de horas
		entry := models.TimeBalanceEntry{
			Amount:      seconds, // Usa os segundos diretamente
			Date:        formattedDate,
			EmployeeID:  employeeID,
			Observation: observation,
			Withdraw:    withdraw,
		}

		// Cria o lançamento no banco de horas
		utils.Logger.Printf("Criando lançamento para o funcionário ID: %s (%s), Segundos: %.2f (%.2f horas), Data: %s",
			employeeID, employeeName, seconds, hours, dateStr)
		err = services.CreateTimeBalanceEntry(b.config, entry)
		if err != nil {
			errorMsg := fmt.Sprintf("Linha %d (%s): %v", i, employeeName, err)
			utils.Logger.Println(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			errorCount++
		} else {
			successCount++
			utils.Logger.Printf("Lançamento criado com sucesso para o funcionário ID: %s (%s)", employeeID, employeeName)
		}

		// Pequena pausa para não sobrecarregar a API
		time.Sleep(500 * time.Millisecond)
	}

	// Prepara a mensagem de resultado
	var resultText strings.Builder
	resultText.WriteString(fmt.Sprintf("Processamento concluído!\n\nLançamentos criados com sucesso: %d\nErros: %d\n", successCount, errorCount))

	// Adiciona detalhes dos erros, se houver
	if errorCount > 0 {
		resultText.WriteString("\nDetalhes dos erros:\n")
		// Limita a quantidade de erros mostrados para não exceder o limite de mensagem do Telegram
		maxErrorsToShow := 10
		if len(errorDetails) > maxErrorsToShow {
			for i := 0; i < maxErrorsToShow; i++ {
				resultText.WriteString("- " + errorDetails[i] + "\n")
			}
			resultText.WriteString(fmt.Sprintf("... e mais %d erros.\n", len(errorDetails)-maxErrorsToShow))
		} else {
			for _, errDetail := range errorDetails {
				resultText.WriteString("- " + errDetail + "\n")
			}
		}
	}

	// Envia a mensagem com o resultado
	resultMsg := tgbotapi.NewMessage(message.Chat.ID, resultText.String())
	b.api.Send(resultMsg)
}
