package main

import (
	"fmt"

	"github.com/jeffemart/PontoGo/app/internal/config"
	"github.com/jeffemart/PontoGo/app/internal/models"
	"github.com/jeffemart/PontoGo/app/internal/services/telegram"
	"github.com/jeffemart/PontoGo/app/internal/utils"
)

// Variável global para armazenar as configurações
var cfg *models.Config

// init é chamado automaticamente antes da função main
func init() {
	// Inicializa o logger
	utils.InitLogger()
	utils.Logger.Println("Logger inicializado com sucesso")

	// Carregar as configurações
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		utils.Logger.Fatalf("Erro ao carregar as configurações: %v", err)
	}
	utils.Logger.Println("Configurações carregadas com sucesso")
}

func main() {
	// Exibir as configurações carregadas
	fmt.Println("Configurações carregadas:")
	fmt.Println("PontoMais Token:", cfg.PontoMaisToken)
	fmt.Println("PontoMais Base URL:", cfg.PontoMaisBaseURL)
	fmt.Println("Telegram Bot Token:", cfg.TelegramBotToken)
	fmt.Println("Telegram Hosts:", cfg.TelegramHosts)
	fmt.Println("Debug:", cfg.Debug)

	// Verificar se o token do bot está definido
	if cfg.TelegramBotToken == "" {
		utils.Logger.Fatal("Erro: TELEGRAM_BOT_TOKEN não definido.")
	}

	// Inicializa o bot do Telegram
	bot, err := telegram.NewBot(cfg)
	if err != nil {
		utils.Logger.Fatalf("Erro ao inicializar o bot do Telegram: %v", err)
	}
	utils.Logger.Println("Bot do Telegram inicializado com sucesso")

	// Inicia o bot
	utils.Logger.Println("Iniciando o bot do Telegram...")
	bot.Start()
}
