package services

import (
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
