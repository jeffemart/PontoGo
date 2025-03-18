package main

import (
	"fmt"
	"log"

	"github.com/jeffemart/PontoGo/app/internal/config"
	"github.com/jeffemart/PontoGo/app/internal/services/telegram"
)

func main() {
	// Carregar as configurações
	cfg := config.LoadConfig()

	// Exibir as configurações carregadas
	fmt.Println("Configurações carregadas:")
	fmt.Println("PontoMais Token:", cfg.PontoMaisToken)
	fmt.Println("PontoMais Base URL:", cfg.PontoMaisBaseURL)
	fmt.Println("Telegram Bot Token:", cfg.TelegramBotToken)
	fmt.Println("Telegram Hosts:", cfg.TelegramHosts)
	fmt.Println("Debug:", cfg.Debug)

	// Verificar se o token do bot está definido
	if cfg.TelegramBotToken == "" {
		log.Fatal("Erro: TELEGRAM_BOT_TOKEN não definido.")
	}

	// Iniciar o bot do Telegram
	telegram.StartBot(cfg)
}
