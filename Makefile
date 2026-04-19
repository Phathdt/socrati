.PHONY: run build clean deps docker-up docker-down dev dev-setup format format-check install-hooks uninstall-hooks test embed

# Source .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# --- Run targets ---

run: build
	./bin/socrati serve

build:
	go build -o bin/socrati .

clean:
	rm -rf bin/

# --- Tests ---

test:
	go test ./... -count=1

# --- Embedder smoke test ---
# Usage: make embed TEXT="hello world"
TEXT ?= hello

embed: build
	./bin/socrati embed --text "$(TEXT)"

# --- Docker ---

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

# --- Dependencies ---

deps:
	go mod tidy

# --- Formatting ---

format:
	@echo "🎨 Formatting all Go files..."
	@gofmt -w .
	@echo "📦 Organizing imports..."
	@goimports -w .
	@echo "📏 Formatting line lengths..."
	@golines -w -m 120 .
	@echo "✨ Applying gofumpt formatting..."
	@gofumpt -extra -w .
	@echo "📄 Formatting YAML files..."
	@npx prettier --write "*.yml" "*.yaml" "docker-compose.yml" 2>/dev/null || true
	@echo "✅ All files formatted successfully!"

format-check:
	@gofmt -l . | grep -q . && echo "Run 'make format' to fix" && exit 1 || echo "✅ Go formatted"
	@npx prettier --check "*.yml" "*.yaml" "docker-compose.yml" 2>/dev/null || (echo "Run 'make format' to fix YAML" && exit 1)

# --- Development ---

dev: docker-up
	@sleep 3
	@make run

dev-setup: format deps install-hooks

# --- Git hooks (lefthook) ---

install-hooks:
	@command -v lefthook >/dev/null || { echo "lefthook not found. Install: brew install lefthook"; exit 1; }
	@lefthook install
	@echo "✅ lefthook hooks installed"

uninstall-hooks:
	@command -v lefthook >/dev/null || { echo "lefthook not found"; exit 0; }
	@lefthook uninstall
	@echo "✅ lefthook hooks removed"
