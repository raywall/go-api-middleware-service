#!/bin/bash

# Script para iniciar os serviços com Docker Compose, validando a chave da API do Datadog

# Obtém o diretório onde o script está localizado
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"

# Função para exibir uso do script
usage() {
    echo "Uso: $0 [DD_API_KEY]"
    echo "  DD_API_KEY: Chave da API do Datadog (obrigatória)"
    echo "  Passe a chave como argumento ou defina-a como variável de ambiente (DD_API_KEY_ENV)"
    exit 1
}

# Função para verificar se o Docker Compose está instalado
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        echo "Erro: Docker Compose não está instalado. Por favor, instale o Docker Compose primeiro."
        exit 1
    fi
}

# Função para verificar se a chave da API foi fornecida
check_api_key() {
    if [ -z "$DD_API_KEY" ]; then
        echo "Erro: A chave da API do Datadog (DD_API_KEY) é obrigatória."
        echo "Passe a chave como argumento ou defina-a como variável de ambiente (DD_API_KEY_ENV)."
        usage
    fi
}

# Função para verificar se o arquivo docker-compose.yml existe
check_compose_file() {
    echo "Verificando arquivo: $COMPOSE_FILE"
    if [ ! -f "$COMPOSE_FILE" ]; then
        echo "Erro: O arquivo $COMPOSE_FILE não foi encontrado."
        exit 1
    fi
}

# Verifica os argumentos
if [ $# -eq 1 ]; then
    DD_API_KEY="$1"
elif [ -n "$DD_API_KEY_ENV" ]; then
    DD_API_KEY="$DD_API_KEY_ENV"
else
    usage
fi

# Verifica o Docker Compose e o arquivo
check_docker_compose
check_compose_file

# Exporta a variável DD_API_KEY para o ambiente do docker-compose
export DD_API_KEY="$DD_API_KEY"

# Inicia os serviços com Docker Compose
echo "Iniciando serviços (Datadog, MySQL, Redis) com arquivo $COMPOSE_FILE..."
docker-compose -f "$COMPOSE_FILE" up -d

if [ $? -eq 0 ]; then
    echo "Serviços iniciados com sucesso!"
    echo " - Datadog: localhost:8126"
    echo " - MySQL: localhost:3306 (user: app_user, password: app_password, database: pricing_db)"
    echo " - Redis: localhost:6379"
else
    echo "Erro ao iniciar os serviços. Verifique os logs com 'docker-compose -f $COMPOSE_FILE logs'."
    exit 1
fi