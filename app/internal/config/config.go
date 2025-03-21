package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jeffemart/PontoGo/app/internal/models"
	"github.com/joho/godotenv"
)

func LoadConfig() (*models.Config, error) {
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
				return nil, fmt.Errorf("erro ao converter TELEGRAM_HOSTS: %v", err)
			}
		}
	}

	// Converter DEBUG para booleano
	debug, erro := strconv.ParseBool(os.Getenv("DEBUG"))
	if erro != nil {
		debug = false
	}

	// Criar a configuração
	cfg := &models.Config{
		PontoMaisToken:   os.Getenv("PONTOMAIS_TOKEN"),
		PontoMaisBaseURL: os.Getenv("PONTOMAIS_BASE_URL"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramHosts:    telegramHosts,
		Debug:            debug,
	}

	// Validação básica das configurações
	if cfg.PontoMaisToken == "" {
		return nil, fmt.Errorf("PONTOMAIS_TOKEN não definido")
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN não definido")
	}

	if len(cfg.TelegramHosts) == 0 {
		return nil, fmt.Errorf("TELEGRAM_HOSTS não definido ou inválido")
	}

	// Configuração padrão para a URL base se não estiver definida
	if cfg.PontoMaisBaseURL == "" {
		cfg.PontoMaisBaseURL = "https://api.pontomais.com.br/external_api/v1"
	}

	return cfg, nil
}
