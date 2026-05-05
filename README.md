# tg_bot_golang

Telegram bot that saves links from users and periodically sends one random saved link back after every `SEND_EVERY_N` saved links.

## How It Works

- User sends a full `http://` or `https://` URL.
- The bot validates and saves the link for that Telegram user.
- Duplicate links are ignored and reported to the user.
- The bot increments the user's saved-link counter.
- If the counter is divisible by `SEND_EVERY_N`, the bot sends a random saved link.
- `/rnd` sends a random saved link manually.
- `/start` and `/help` show help text.

## Architecture

The project keeps the existing package layout:

- `app/` - application bootstrap, dependency wiring, logger, graceful shutdown.
- `clients/telegram` - Telegram HTTP client.
- `consumer/poller` - event polling orchestration.
- `events/telegram` - Telegram event parsing and command routing.
- `service/` - business logic for links, counters, URL validation, random-send decision.
- `storage/` - storage interface plus sqlite and files implementations.
- `config/` - environment configuration.
- `lib/e` - small error wrapping helpers.

More detail is in [ARCHITECTURE.md](ARCHITECTURE.md).

## Configuration

Create `.env` from `.env.example` or export variables directly:

```bash
cp .env.example .env
```

Required variables:

- `BOT_TOKEN` - Telegram bot token from BotFather.
- `DATABASE_PATH` - sqlite database file path, for example `data/sqlite/storage.db`.
- `SEND_EVERY_N` - send a random saved link after every N saved links.
- `LOG_LEVEL` - `debug`, `info`, `warn`, or `error`.
- `ENV` - `local`, `dev`, `staging`, or `production`.

`DATABASE_URL` is accepted as a fallback for `DATABASE_PATH`.
`TELEGRAM_HOST` is optional and defaults to `api.telegram.org`. It may also be set to a full base URL for testing through a local proxy or mock server.

## Local Run

Export env variables, then run:

```bash
go run .
```

Or:

```bash
make run
```

The app creates the sqlite directory and schema on startup.

## Tests

```bash
go test ./...
go test -race ./...
```

Convenience target:

```bash
make test
```

## Quality Checks

```bash
test -z "$(gofmt -l .)"
go vet ./...
```

Or:

```bash
make lint
```

## Storage

Production storage is sqlite. It creates:

- `links` with `user_id`, `link`, `created_at`, and a unique `(user_id, link)` constraint.
- `message_counters` with `user_id`, `count`, and `updated_at`.

The files storage remains available as a lightweight implementation of the storage interface, but sqlite is the recommended production backend.

## Structure

- `main.go` stays thin and delegates bootstrap into `app.Run()`.
- Import paths are lowercase and package names avoid underscores and hyphens.
- Runtime dependencies are wired once in `app/`, while transport and business logic remain isolated.

## CI

GitHub Actions workflow is in `.github/workflows/ci.yml` and runs:

- `go mod download`
- `gofmt` check
- `go vet ./...`
- `go test ./...`
- `go test -race ./...`

## Documentation

- Static docs site source is in `docs/`.
- GitHub Pages deployment workflow is in `.github/workflows/pages.yml`.
- After enabling Pages with the `GitHub Actions` source in repository settings, pushes to `main` publish the docs site automatically.

## Troubleshooting

- `BOT_TOKEN is required`: export `BOT_TOKEN` or load `.env` before running.
- `DATABASE_PATH or DATABASE_URL is required`: set a sqlite path.
- `duplicate link`: the bot ignores repeated links for the same user.
- No random link is sent: verify `SEND_EVERY_N` and that the user has saved links.
- sqlite build errors: `github.com/mattn/go-sqlite3` uses CGO, so install a C compiler in the runtime/build image.
