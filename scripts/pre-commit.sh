#!/usr/bin/env bash
# Pre-commit hook: format staged Go files and re-stage them.
# Fails the commit if any formatter fails.
set -euo pipefail

# Only look at staged .go files (Added, Copied, Modified, Renamed).
# Portable: avoid bash 4+ `mapfile` since macOS ships bash 3.2.
FILES=()
while IFS= read -r line; do
    FILES+=("$line")
done < <(git diff --cached --name-only --diff-filter=ACMR | grep -E '\.go$' || true)

if [[ ${#FILES[@]} -eq 0 ]]; then
    exit 0
fi

echo "▶ pre-commit: formatting ${#FILES[@]} Go file(s)"

run_if_installed() {
    local bin="$1"
    shift
    if command -v "$bin" >/dev/null 2>&1; then
        "$bin" "$@"
    fi
}

run_if_installed gofmt -w "${FILES[@]}"
run_if_installed goimports -w "${FILES[@]}"
run_if_installed gofumpt -extra -w "${FILES[@]}"
run_if_installed golines -w -m 120 "${FILES[@]}"

# Re-stage the (possibly) reformatted files so the commit includes changes.
git add -- "${FILES[@]}"

# Run go vet on the whole module — catches obvious bugs before commit.
if command -v go >/dev/null 2>&1; then
    echo "▶ pre-commit: go vet ./..."
    go vet ./...
fi

echo "✓ pre-commit: ok"
