# socrati

Minimal Go HTTP service template.

- **Fiber v2** for HTTP
- **urfave/cli/v2** for command routing
- **viper** YAML config + `__` env overrides
- Custom **slog** logger with colored text output (`pkg/logger`)
- Built-in `request_id` + `latency_ms` request logging
- `/health` endpoint

## Quick start

```bash
cp config.yml.example config.yml
make run
# server up on http://localhost:4000

curl -s http://127.0.0.1:4000/health
# {"status":"ok"}
```

## Layout

```
main.go                  CLI bootstrap
cli/serve_svc.go         `serve` command
cmd/httpapi/             Fiber app, /health, request middleware
config/                  Viper loader
pkg/logger/              Logger interface (wraps slog)
```

## Configuration

Edit `config.yml` (gitignored). Override via env using `__` as separator:

| Env var          | Key             | Default       |
| ---------------- | --------------- | ------------- |
| `SERVER__HOST`   | `server.host`   | `localhost`   |
| `SERVER__PORT`   | `server.port`   | `4000`        |
| `LOGGER__LEVEL`  | `logger.level`  | `info`        |
| `LOGGER__FORMAT` | `logger.format` | `text`        |
| `ENV`            | `env`           | `development` |

Logger formats: `text` (colored), `json`, `plain` (no color).

## Commands

```bash
make build          # → bin/socrati
make run            # build + serve
make deps           # go mod tidy
make format         # gofmt + goimports + golines + gofumpt
make docker-up      # docker compose up --build -d
make docker-down    # docker compose down
make install-hooks  # install lefthook pre-commit hooks
go test ./...       # run tests
```

Run binary directly:

```bash
./bin/socrati serve --config config.yml
```

## License

MIT
