package telegram

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jeffemart/PontoGo/app/internal/models"
	services "github.com/jeffemart/PontoGo/app/internal/services/pontomais"
)

// StartBot inicializa e executa o bot do Telegram
func StartBot(cfg *models.Config) {
	// Cria o bot com o token
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Erro ao criar o bot: %v", err)
	}

	bot.Debug = cfg.Debug
	log.Printf("Bot autorizado com %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatalf("Erro ao obter atualizações: %v", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		handleMessage(bot, update, cfg)
	}
}

// handleMessage processa as mensagens recebidas
func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, cfg *models.Config) {
	var msg tgbotapi.MessageConfig
	msg.ChatID = update.Message.Chat.ID
	msg.ParseMode = "Markdown"

	switch update.Message.Text {
	case "/start":
		msg.Text = "✨ *Bem-vindo ao nosso bot!* ✨\n\nUse /ajuda para ver os comandos disponíveis."
	case "/ajuda":
		msg.Text = "📌 *Comandos disponíveis:*\n\n🔹 `/start` - Iniciar o bot\n🔹 `/ajuda` - Mostrar ajuda\n🔹 `/funcionarios` - Listar funcionários"
	case "/funcionarios":
		employees, err := services.GetEmployees(cfg)
		log.Printf("Número de funcionários recebidos: %d", len(employees))
		
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Erro ao obter os funcionários.")
			bot.Send(msg)
			break
		}

		var response string
		if len(employees) == 0 {
			response = "Nenhum funcionário encontrado."
		} else {
			// Retorna apenas a quantidade de funcionários
			response = fmt.Sprintf("Total de funcionários: %d", len(employees))
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	default:
		msg.Text = "📌 *Comandos disponíveis:*\n\n🔹 `/start` - Iniciar o bot\n🔹 `/ajuda` - Mostrar ajuda\n🔹 `/funcionarios` - Listar funcionários"
	}

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
	}
}
