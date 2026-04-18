.PHONY: run build clean deps docker-up docker-down dev dev-setup format format-check install-hooks uninstall-hooks

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

# --- Git hooks ---

HOOK_SRC := scripts/pre-commit.sh
HOOK_DST := .git/hooks/pre-commit

install-hooks:
	@if [ ! -d .git ]; then echo "not a git repo"; exit 1; fi
	@mkdir -p .git/hooks
	@ln -sf ../../$(HOOK_SRC) $(HOOK_DST)
	@chmod +x $(HOOK_SRC)
	@echo "✅ pre-commit hook installed → $(HOOK_DST)"

uninstall-hooks:
	@rm -f $(HOOK_DST)
	@echo "✅ pre-commit hook removed"
