# PontoGo

PontoGo é uma biblioteca Go que integra a API do Ponto Mais e o Telegram, permitindo o gerenciamento e edição do banco de horas de funcionários através de um bot no Telegram. Através do bot, os administradores podem visualizar e atualizar os horários de trabalho dos funcionários, além de enviar planilhas Excel com os valores do banco de horas.

## Funcionalidades

- Integração com a **API do Ponto Mais** para acessar e editar os dados de banco de horas dos funcionários.
- **Bot do Telegram** para interagir com os dados, editar e atualizar informações diretamente através de mensagens.
- Envio de **planilhas Excel** contendo os dados do banco de horas para os administradores.
- Permite que os administradores atualizem as horas de trabalho de forma automatizada e eficiente.

## Instalação

### Requisitos
Certifique-se de que você tenha os seguintes requisitos instalados:
- [Go](https://golang.org/dl/) (versão 1.18 ou superior)
- Conta no [Telegram](https://telegram.org/) e um bot criado através do [BotFather](https://core.telegram.org/bots#botfather).
- Conta no [Ponto Mais](https://www.pontomais.com.br/) para integração com a API.

### Usando Docker (Recomendado)

1. Clone o repositório:
   ```bash
   git clone https://github.com/jeffemart/PontoGo.git
   cd PontoGo
   ```

2. Copie o arquivo de exemplo de ambiente e configure suas variáveis:
   ```bash
   cp .env_example .env
   ```

3. Execute com Docker Compose:
   ```bash
   docker-compose up -d
   ```

### Instalação Manual

1. Clone o repositório:
   ```bash
   git clone https://github.com/jeffemart/PontoGo.git
   cd PontoGo
   ```

2. Copie o arquivo de exemplo de ambiente e configure suas variáveis:
   ```bash
   cp .env_example .env
   ```

3. Instale as dependências:
   ```bash
   go mod download
   ```

4. Execute a aplicação:
   ```bash
   go run app/cmd/main.go
   ```

## Configuração

Configure as seguintes variáveis de ambiente no arquivo `.env`:

```env
# Configurações da API do Ponto Mais
PONTOMAIS_TOKEN="seu_token"
PONTOMAIS_BASE_URL="https://url.dominio.com"

# Configurações do Bot do Telegram
TELEGRAM_BOT_TOKEN="seu_bot_token"
TELEGRAM_HOSTS=123456789,987654321

# Modo Debug
DEBUG=false
```

## Docker Hub

A imagem Docker está disponível no Docker Hub:

```bash
docker pull <seu_usuario>/pontogo:latest
```

## Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.
