# Makefile for the pricing microservice
# Provides commands to build, run, test, benchmark, and clean the project

# Variáveis
COMPOSE_FILE := config/docker-compose.yml
RUN_SERVICES_SCRIPT := config/run-services.sh
DD_API_KEY ?= "datadog-key" # $(DD_API_KEY_ENV)

# Verifica se o script run-services.sh existe
CHECK_SCRIPT := $(shell if [ -f "$(RUN_SERVICES_SCRIPT)" ]; then echo "found"; else echo "not-found"; fi)

# Variables
BINARY_NAME=pricing-service
GO=go
GOFLAGS=-v

# Default target
.PHONY: all
all: build

# Verifica pré-requisitos
.PHONY: check
check:
	@if [ "$(CHECK_SCRIPT)" = "not-found" ]; then \
		echo "Erro: O script $(RUN_SERVICES_SCRIPT) não foi encontrado."; \
		exit 1; \
	fi
	@if [ -z "$(DD_API_KEY)" ]; then \
		echo "Erro: A variável DD_API_KEY ou DD_API_KEY_ENV deve ser definida."; \
		echo "Exemplo: make DD_API_KEY=your-api-key services"; \
		exit 1; \
	fi
	@if ! command -v docker-compose >/dev/null 2>&1; then \
		echo "Erro: Docker Compose não está instalado."; \
		exit 1; \
	fi
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "Erro: Docker não está instalado."; \
		exit 1; \
	fi

# Inicia os serviços (Datadog, MySQL, Redis)
.PHONY: services
services: check
	@echo "Iniciando serviços com $(RUN_SERVICES_SCRIPT)..."
	@chmod +x $(RUN_SERVICES_SCRIPT)
	@$(RUN_SERVICES_SCRIPT) $(DD_API_KEY)

# Para e remove os serviços
.PHONY: stop
stop:
	@echo "Parando e removendo serviços..."
	@docker compose -f $(COMPOSE_FILE) down

# Verifica o status dos serviços
.PHONY: status
status:
	@echo "Verificando status dos serviços..."
	@docker compose -f $(COMPOSE_FILE) ps

# Build the project and generate the binary
.PHONY: build
build:
	@echo "Compilando o código Go..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) main.go

# Run the project directly
.PHONY: run
run:
	@echo "Executando a aplicação Go..."
	@DD_ENABLED=true $(GO) run $(GOFLAGS) .

# Run all tests with verbose output
.PHONY: test
test:
	@echo "Executando testes com verbose output..."
	@$(GO) test ./... $(GOFLAGS) > test_result.out

# Run all benchmarks with memory profiling
.PHONY: bench
bench:
	@echo "Executando benchmarks com memória de perfil..."
	$(GO) test -bench=. -benchmem ./... > bench_result.out

# Clean up generated files
.PHONY: clean
clean:
	@echo "Parando e removendo serviços e volumes..."
	@docker compose -f $(COMPOSE_FILE) down -v
	@rm -f $(BINARY_NAME)

# Ensure dependencies are installed
.PHONY: deps
deps:
	@echo "Instalando dependências..."
	@$(GO)fmt -w .
	@$(GO) mod tidy