# PontoGo

PontoGo é uma aplicação em Go que integra a API do Ponto Mais com o Telegram, permitindo o gerenciamento e edição do banco de horas de funcionários através de um bot. Através do bot, os administradores podem visualizar e atualizar os horários de trabalho dos funcionários de forma eficiente.

## Funcionalidades do Bot

### Comandos Disponíveis
- `/start` - Inicia o bot e exibe mensagem de boas-vindas
- `/help` - Mostra a lista de comandos disponíveis
- `/listar` - Lista todos os colaboradores ativos
- `/editar` - Edita um lançamento existente no banco de horas
- `/criar` - Cria um novo lançamento no banco de horas

### Exemplos de Uso

#### Editar Lançamento
```bash
/editar <ID> <quantidade_segundos> <data> <observação> <retirada>

# Exemplo: Adicionar 1 minuto (60 segundos)
/editar 3833376 60.0 2025-03-18 "Ajuste de ponto" false
```

#### Criar Lançamento
```bash
/criar <ID_funcionário> <quantidade_segundos> <data> <observação> <retirada>

# Exemplo: Registrar 2 horas (7200 segundos)
/criar 1487972 7200.0 2025-03-18 "Horas extras" false
```

## Instalação

### Requisitos
- [Go](https://golang.org/dl/) (versão 1.18 ou superior) - apenas para instalação manual
- [Docker](https://docs.docker.com/get-docker/) (para execução via container)
- Conta no [Telegram](https://telegram.org/) e um bot criado através do [BotFather](https://core.telegram.org/bots#botfather)
- Conta no [Ponto Mais](https://www.pontomais.com.br/) para integração com a API

### Usando Docker (Recomendado)

1. Pull da imagem oficial:
```bash
docker pull <seu_usuario>/pontogo:latest
```

2. Copie e configure o arquivo de ambiente:
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

2. Configure o ambiente:
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

### Variáveis de Ambiente
Configure as seguintes variáveis no arquivo `.env`:

```env
# Configurações da API do Ponto Mais
PONTOMAIS_TOKEN="seu_token"
PONTOMAIS_BASE_URL="https://api.pontomais.com.br/external_api/v1"

# Configurações do Bot do Telegram
TELEGRAM_BOT_TOKEN="seu_bot_token"
TELEGRAM_HOSTS=123456789,987654321  # IDs dos chats autorizados

# Modo Debug
DEBUG=false
```

### Configuração do Docker
O arquivo `docker-compose.yml` já está configurado com:
- Reinício automático do container
- Volume para o arquivo .env
- Configurações de ambiente

## Segurança

- Apenas usuários com chat IDs autorizados podem interagir com o bot
- As credenciais são gerenciadas via variáveis de ambiente
- O container Docker executa com privilégios mínimos

## CI/CD

O projeto utiliza GitHub Actions para:
- Build automático da imagem Docker
- Push para o Docker Hub em cada commit na main
- Versionamento automático das imagens

## Docker Hub

A imagem oficial está disponível no Docker Hub:
```bash
docker pull <seu_usuario>/pontogo:latest
```

## Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.
