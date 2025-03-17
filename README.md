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

### Passos para Instalação

1. Clone o repositório:
   ```bash
   git clone https://github.com/jeffemart/PontoGo.git
   cd PontoGo
