# Makefile

# Переменные
APP_NAME = sap_segmentationd
CMD_DIR = ./cmd/sap_segmentationd
BINARY = $(APP_NAME)

# Цвета для вывода
GREEN = \033[0;32m
NC = \033[0m # No Color

.PHONY: all build lint run

# По умолчанию выполняется сборка
all: build

build:
	@echo "$(GREEN)Building $(APP_NAME)...$(NC)"
	@go build -o $(BINARY) $(CMD_DIR)/main.go

lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@golangci-lint run ./...

run: build
	@echo "$(GREEN)Running $(APP_NAME)...$(NC)"
	@./$(BINARY)
