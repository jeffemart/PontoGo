# PontoGo

PontoGo é uma aplicação em Go que integra a API do Ponto Mais com o Telegram, permitindo o gerenciamento e edição do banco de horas de funcionários através de um bot. Através do bot, os administradores podem visualizar e atualizar os horários de trabalho dos funcionários de forma eficiente.

## Funcionalidades do Bot

### Comandos Disponíveis
- `/start` - Inicia o bot e exibe mensagem de boas-vindas
- `/help` - Mostra a lista de comandos disponíveis
- `/listar` - Lista todos os colaboradores ativos
- `/editar` - Edita um lançamento existente no banco de horas
- `/criar` - Cria um novo lançamento no banco de horas
- `/relatorio` - Processa um arquivo Excel para criar múltiplos lançamentos no banco de horas

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

#### Processar Relatório em Lote
O comando `/relatorio` permite processar múltiplos lançamentos de banco de horas a partir de um arquivo Excel.

1. Envie o comando `/relatorio`
2. Envie um arquivo Excel (.xlsx) com as seguintes colunas:
   - **ID**: ID do funcionário no sistema Ponto Mais
   - **NOME**: Nome do funcionário (para referência)
   - **DATA**: Data do lançamento no formato DD/MM/AAAA ou AAAA-MM-DD
   - **HORAS**: Quantidade em segundos a ser lançada
   - **OBSERVAÇÃO**: Descrição/motivo do lançamento
   - **DEBITO**: Indicador se é uma retirada (TRUE/FALSE, SIM/NÃO)

3. O bot processará cada linha do arquivo e criará os lançamentos correspondentes no banco de horas.

**Exemplo de arquivo Excel:**
| ID       | NOME                           | DATA       | HORAS | OBSERVAÇÃO        | DEBITO |
|----------|--------------------------------|------------|-------|-------------------|--------|
| 1234567  | NOME DO COLABORADOR 1          | 19/03/2025 | 36000 | Horas Extras Pagas| TRUE   |
| 7654321  | NOME DO COLABORADOR 2          | 19/03/2025 | 38460 | Horas Extras Pagas| TRUE   |

**Observações:**
- A primeira linha do arquivo deve conter os cabeçalhos
- O valor em HORAS deve ser em segundos (ex: 3600 = 1 hora)
- O campo DEBITO aceita TRUE/FALSE, SIM/NÃO, S/N (não sensível a maiúsculas/minúsculas)

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
