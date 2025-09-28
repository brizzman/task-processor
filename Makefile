.PHONY: dev dev-down check lint security test tools help


# ----------------------------
# Load environment variables 
# ----------------------------
ENV_FILE := .env.dev
ifneq ("$(wildcard $(ENV_FILE))","")
    include $(ENV_FILE)
    export
endif


# ----------------------------
# HELP: show available commands
# ----------------------------
help:
	@echo "================ Available Targets ================"
	@echo "  dev       - Start infrastructure"
	@echo "  dev-down  - Stop infrastructure"
	@echo "  run       - Run the application"
	@echo "  check     - Run Linting, Security scanning, and Tests"
	@echo "  lint      - Run code linter"
	@echo "  security  - Run security scanning"
	@echo "  test      - Run tests"
	@echo "  tools     - Install required tools"
	@echo "  help      - Show this help message"
	@echo "==================================================="


# ----------------------------
# DEV: Infrastructure
# ----------------------------
dev:
	@echo "Starting infrastructure..."
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Waiting for Postgres..."
	./scripts/wait-for-it.sh localhost:5433 --timeout=30 --strict -- echo "Postgres is ready"
	@echo "Waiting for Redis..."
	./scripts/wait-for-it.sh localhost:6380 --timeout=30 --strict -- echo "Redis is ready"
	@echo "Infrastructure started."

dev-down:
	@echo "Stopping infrastructure..."
	docker-compose -f docker-compose.dev.yml down --volumes --remove-orphans
	@echo "Infrastructure stopped."


# ----------------------------
# RUN: Start the application
# ----------------------------
run:
	@echo "Running application..."
	go run ./cmd/app/main.go


# ----------------------------
# CHECK: Full check (lint + security + tests)
# ----------------------------
check: lint security test


# ----------------------------
# Linting
# ----------------------------
lint:
	@echo "Running Linting..."
	golangci-lint run


# ----------------------------
# Security scanning
# ----------------------------
security:
	@echo "Running Security Scanning..."
	gosec ./...


# ----------------------------
# Tests
# ----------------------------
test:
	@echo "Running Tests..."
	go test ./...

# ----------------------------
# TOOLS: Install required tools
# ----------------------------
tools:
	@echo "Installing required tools..."
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed."
