package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jeffemart/PontoGo/app/internal/models"
)

// GetEmployees faz a requisição à API do Ponto Mais para listar colaboradores
func GetEmployees(cfg *models.Config) ([]models.Employee, error) {
	if cfg.PontoMaisBaseURL == "" || cfg.PontoMaisToken == "" {
		log.Println("Erro: Variáveis de ambiente PONTOMAIS_TOKEN ou PONTOMAIS_BASE_URL não estão definidas.")
		return nil, fmt.Errorf("variáveis de ambiente não definidas corretamente")
	}

	// Monta a URL correta utilizando cfg.PontoMaisBaseURL
	url := fmt.Sprintf("%s/employees?active=true&attributes=id,first_name,last_name,email,cpf,registration_number&sort_direction=asc&sort_property=first_name", cfg.PontoMaisBaseURL)

	// Cria a requisição HTTP
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Erro ao criar a requisição: %v", err)
		return nil, err
	}

	// Adiciona o cabeçalho com o token de autenticação
	req.Header.Add("access-token", cfg.PontoMaisToken)

	// Executa a requisição
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao realizar a requisição: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Lê o corpo da resposta para debugging
		log.Printf("Erro na resposta da API. Status: %s. Resposta: %s", resp.Status, string(body))
		return nil, fmt.Errorf("erro na resposta da API, status: %s", resp.Status)
	}

	// Lê e processa o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler o corpo da resposta: %v", err)
		return nil, err
	}

	// log.Printf("Resposta da API do Ponto Mais: %s", string(body))

	// Estrutura para armazenar a resposta
	var result models.EmployeesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Erro ao deserializar os dados: %v", err)
		return nil, err
	}

	return result.Employees, nil
}

// UpdateTimeBalanceEntry atualiza o banco de horas de um funcionário
func UpdateTimeBalanceEntry(cfg *models.Config, entryID string, entry models.TimeBalanceEntry) error {
	if cfg.PontoMaisBaseURL == "" || cfg.PontoMaisToken == "" {
		log.Println("Erro: Variáveis de ambiente PONTOMAIS_TOKEN ou PONTOMAIS_BASE_URL não estão definidas.")
		return fmt.Errorf("variáveis de ambiente não definidas corretamente")
	}

	// Monta a URL para a requisição
	url := fmt.Sprintf("%s/time_balance_entries/%s", cfg.PontoMaisBaseURL, entryID)

	// Prepara o corpo da requisição
	requestBody := map[string]models.TimeBalanceEntry{
		"time_balance_entry": entry,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Erro ao serializar os dados: %v", err)
		return err
	}

	// Cria a requisição HTTP
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Erro ao criar a requisição: %v", err)
		return err
	}

	// Adiciona os cabeçalhos necessários
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("access-token", cfg.PontoMaisToken)

	// Executa a requisição
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao realizar a requisição: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Verifica o status da resposta
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Erro na resposta da API. Status: %s. Resposta: %s", resp.Status, string(body))
		return fmt.Errorf("erro na resposta da API, status: %s", resp.Status)
	}

	log.Printf("Banco de horas atualizado com sucesso para o ID: %s", entryID)
	return nil
}

// CreateTimeBalanceEntry cria um novo lançamento no banco de horas de um funcionário
func CreateTimeBalanceEntry(cfg *models.Config, entry models.TimeBalanceEntry) error {
	if cfg.PontoMaisBaseURL == "" || cfg.PontoMaisToken == "" {
		log.Println("Erro: Variáveis de ambiente PONTOMAIS_TOKEN ou PONTOMAIS_BASE_URL não estão definidas.")
		return fmt.Errorf("variáveis de ambiente não definidas corretamente")
	}

	// Monta a URL para a requisição
	url := fmt.Sprintf("%s/time_balance_entries", cfg.PontoMaisBaseURL)

	// Prepara o corpo da requisição
	requestBody := map[string]models.TimeBalanceEntry{
		"time_balance_entry": entry,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Erro ao serializar os dados: %v", err)
		return err
	}

	// Cria a requisição HTTP
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Erro ao criar a requisição: %v", err)
		return err
	}

	// Adiciona os cabeçalhos necessários
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("access-token", cfg.PontoMaisToken)

	// Executa a requisição
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao realizar a requisição: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Verifica o status da resposta
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Erro na resposta da API. Status: %s. Resposta: %s", resp.Status, string(body))
		return fmt.Errorf("erro na resposta da API, status: %s", resp.Status)
	}

	log.Printf("Lançamento no banco de horas criado com sucesso")
	return nil
}
