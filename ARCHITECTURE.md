# Architecture

This project keeps the original directory structure and adds a small `service/` layer for business logic.

## Runtime Flow

1. `main.go` delegates startup to `app/`, which loads env config, initializes `slog`, opens sqlite, creates the service, and starts the poller.
2. `consumer/poller` polls events in batches and passes each event to the processor.
3. `events/telegram` converts Telegram updates into internal events and routes messages:
   - commands are handled as Telegram commands;
   - non-command text is passed to the link service.
4. `service` validates links, saves them, increments user counters, and decides when to request a random link.
5. `storage` persists links and counters.
6. `clients/telegram` sends HTTP requests to Telegram.

## Package Responsibilities

### `app/`

- Load `BOT_TOKEN`, database path, `SEND_EVERY_N`, log level, and environment.
- Create cancellable context from OS signals.
- Initialize storage schema.
- Wire Telegram client, processor, service, and poller.
- Close storage during shutdown.

### `main.go`

`main.go` is intentionally thin. It delegates startup to `app.Run()` and only converts a fatal error into process exit.

### `consumer/poller`

The poller is orchestration only. It does not know Telegram commands or link rules. It fetches events, logs failures, respects context cancellation, and delegates processing.

### `events/telegram`

This package is an adapter for Telegram. It parses update metadata, routes commands, turns service results into user-facing messages, and calls the Telegram client.

Business decisions such as URL validation, duplicate policy, counters, and scheduled random sends live in `service/`.

### `service/`

The link service owns business rules:

- validate and normalize URLs;
- save links;
- ignore duplicates;
- increment per-user counters;
- send a random link every `SEND_EVERY_N` saved links;
- expose manual random-link retrieval for `/rnd`.

### `storage/`

The storage interface exposes data access only:

- schema initialization;
- save link;
- get random link for a user;
- increment and read user counters;
- close resources.

The sqlite implementation creates tables and indexes on startup. The files implementation remains for compatibility and local experimentation.

### `clients/telegram`

The Telegram client is a thin HTTP wrapper around `getUpdates` and `sendMessage`. It accepts context so shutdown can cancel in-flight requests.

## Error Handling

Errors are wrapped with `%w` so callers can use `errors.Is`. Domain errors include:

- `storage.ErrNoSavedPages`
- `storage.ErrDuplicateLink`
- `service.ErrInvalidURL`

## Shutdown

`app/` uses `signal.NotifyContext`. When `SIGINT` or `SIGTERM` arrives, polling stops, in-flight HTTP requests receive context cancellation, and sqlite is closed.
