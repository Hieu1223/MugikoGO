# ==============================================================================
# Configuration & Environment
# ==============================================================================

# Export project root so Python scripts can resolve absolute paths
export PROJECT_ROOT := $(CURDIR)

# Directory containing Python scripts
SCRIPTS_DIR := $(PROJECT_ROOT)/scripts

# Python executable inside the virtual environment (adjust path if venv is elsewhere)
PYTHON := $(PROJECT_ROOT)/.venv/Scripts/python.exe

# Load .env.local if it exists and export variables to the environment
-include .env.local
export

# Go build caches
GOCACHE ?= $(PROJECT_ROOT)/.cache/go-build
GOMODCACHE ?= $(PROJECT_ROOT)/.cache/go-mod
export GOCACHE
export GOMODCACHE




# ==============================================================================
# Phony Targets
# ==============================================================================
.PHONY: build run run-dev setup setup-dev clean help

# ==============================================================================
# Help
# ==============================================================================
help:
	@echo "Available commands:"
	@echo "  make build      - Build the Go server"
	@echo "  make run        - Build and run the server"
	@echo "  make run-dev    - Run server with auto DB rollback"
	@echo "  make setup      - Initialize DB (schema + seed)"
	@echo "  make clean      - Remove build artifacts and caches"

# ==============================================================================
# Go Server Targets
# ==============================================================================

build:
	go build -o build/main ./cmd/server

run: build
	xcopy /E /I .\asset ".\build\asset"
	./build/main

run-dev: build
	$(PYTHON) $(SCRIPTS_DIR)/run-dev.py

# ==============================================================================
# Database Targets
# ==============================================================================

setup:
	@echo "Running database setup..."
	$(PYTHON) $(SCRIPTS_DIR)/common/run_sql.py -f $(SCRIPTS_DIR)/sql/setup/schema.sql
	$(PYTHON) $(SCRIPTS_DIR)/common/run_sql.py -f $(SCRIPTS_DIR)/sql/setup/seed.sql

# ==============================================================================
# Cleanup
# ==============================================================================

clean:
	if exist build rmdir /s /q build
	if exist .cache rmdir /s /q .cache
	if exist backup rmdir /s /q backup


reset-db:
	$(PYTHON) $(SCRIPTS_DIR)/reset-db.py