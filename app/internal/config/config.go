package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jeffemart/PontoGo/app/internal/models"
	"github.com/joho/godotenv"
)

func LoadConfig() *models.Config {
	// Carregar as variáveis do arquivo .env
	if erro := godotenv.Load(); erro != nil {
		log.Println("Aviso: Não foi possível carregar o arquivo .env, utilizando variáveis de ambiente.")
	}

	// Converter a lista de hosts do Telegram para slice de int64
	telegramHostsStr := os.Getenv("TELEGRAM_HOSTS")
	var telegramHosts []int64
	if telegramHostsStr != "" {
		for _, host := range strings.Split(telegramHostsStr, ",") {
			if h, err := strconv.ParseInt(strings.TrimSpace(host), 10, 64); err == nil {
				telegramHosts = append(telegramHosts, h)
			} else {
				log.Printf("Erro ao converter TELEGRAM_HOSTS: %v", err)
			}
		}
	}

	// Converter DEBUG para booleano
	debug, erro := strconv.ParseBool(os.Getenv("DEBUG"))
	if erro != nil {
		debug = false
	}

	return &models.Config{
		PontoMaisToken: os.Getenv("PONTOMAIS_TOKEN"),
		PontoMaisBaseURL: os.Getenv("PONTOMAIS_BASE_URL"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramHosts: telegramHosts,
		Debug: debug,
	}
}