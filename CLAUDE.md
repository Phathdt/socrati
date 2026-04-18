# CLAUDE.md

Guidance for Claude Code when working in this repo.

## Project

Minimal Go HTTP service template built on **Fiber v2**, with:

- `urfave/cli/v2` command router
- `viper` YAML config + `__` env overrides
- Custom `slog` logger with colored text output (`pkg/logger`)
- Request middleware: `request_id` (UUID) + `latency_ms` per request
- `/health` endpoint

Use it as a starting point; add modules as needed.

## Layout

```
main.go                  CLI bootstrap (urfave/cli)
cli/serve_svc.go         `serve` command — loads config, builds logger + app
cmd/httpapi/server.go    Fiber app, /health, request logger middleware
config/config.go         Viper loader (yaml + env), Validate()
config.yml               Local config (gitignored)
config.yml.example       Template (committed)
pkg/logger/              Logger interface + ColoredTextHandler (wraps slog)
```

`slog` is intentionally hidden inside `pkg/logger` — never import `log/slog`
elsewhere; depend on `logger.Logger` instead.

## Commands

| Command           | Purpose                                |
| ----------------- | -------------------------------------- |
| `make build`      | Build to `bin/socrati`                 |
| `make run`        | Build then run `serve`                 |
| `make deps`       | `go mod tidy`                          |
| `make format`     | gofmt / goimports / golines / gofumpt  |
| `make docker-up`  | `docker compose up --build -d`         |
| `make docker-down`| `docker compose down`                  |
| `make install-hooks`| Install lefthook pre-commit hooks    |

Run directly:

```bash
./bin/socrati serve --config config.yml
```

## Config

YAML at `config.yml`. Defaults: `localhost:4000`, level `debug`, format `text`.

Env overrides use `__` separator (viper):

| Env var          | Overrides         |
| ---------------- | ----------------- |
| `SERVER__HOST`   | `server.host`     |
| `SERVER__PORT`   | `server.port`     |
| `LOGGER__LEVEL`  | `logger.level`    |
| `LOGGER__FORMAT` | `logger.format`   |
| `ENV`            | `env`             |

Logger formats: `text` (colored, default), `json`, `plain` (uncolored).

## Conventions

- Go file naming: `snake_case.go`.
- Don't import `log/slog` outside `pkg/logger`.
- Keep `main.go` thin — wire commands only; logic lives in `cli/` + `cmd/`.
- Prefer editing existing files over creating new ones; YAGNI/KISS/DRY.
- New CLI commands go in `cli/<name>_svc.go` with action `Run<Name>` exported.
- New HTTP routes go in `cmd/httpapi/`.

## Adding a new endpoint

1. Add handler in `cmd/httpapi/` (or new sub-package under `cmd/`).
2. Register in `NewApp()` (or compose a new app builder).
3. Middleware via `app.Use(...)`. `RequestLogger` already wired.

## Testing

```bash
go test ./...
```

Smoke test:

```bash
curl -s http://127.0.0.1:4000/health   # → {"status":"ok"}
```
