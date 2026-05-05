# Architecture

This project keeps the original directory structure and adds a small `service/` layer for business logic.

## Runtime Flow

1. `main.go` delegates startup to `app/`, which loads env config from `os.Getenv`, initializes `slog`, opens sqlite, loads the last persisted Telegram offset, creates the Telegram client and service, and starts the poller.
2. `consumer/poller` polls events in batches and passes each event to the processor.
3. `events/telegram` converts Telegram updates into internal events and routes messages:
   - commands are handled as Telegram commands;
   - non-command text is passed to the link service.
4. `service` validates links, saves them, increments user counters, and decides when to request a random link.
5. `storage` persists links, counters, and internal bot state.
6. `clients/telegram` sends HTTP requests to Telegram.

## Package Responsibilities

### `app/`

- Load `BOT_TOKEN`, database path, `SEND_EVERY_N`, log level, and environment.
- Validate the configured Telegram API host during startup.
- Create cancellable context from OS signals.
- Initialize storage schema.
- Load the last saved Telegram offset from sqlite.
- Wire Telegram client, processor, service, and poller.
- Close storage during shutdown.

### `main.go`

`main.go` is intentionally thin. It delegates startup to `app.Run()` and only converts a fatal error into process exit.

### `consumer/poller`

The poller is orchestration only. It does not know Telegram commands or link rules. It fetches events, logs failures, respects context cancellation, delegates processing, and stops the current batch on the first processing error so failed updates are retried in order.

### `events/telegram`

This package is an adapter for Telegram. It parses update metadata, routes commands, turns service results into user-facing messages, calls the Telegram client, and persists the next Telegram offset only after an update has been processed successfully.

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
- load and save internal Telegram offset state;
- close resources.

The sqlite implementation creates tables and indexes on startup, including `bot_state` for runtime metadata such as the Telegram offset. The files implementation remains for compatibility and local experimentation.

### `clients/telegram`

The Telegram client is a thin HTTP wrapper around `getUpdates` and `sendMessage`. It validates the configured base URL during construction, uses `GET` for update polling, uses `POST` form bodies for `sendMessage`, and accepts context so shutdown can cancel in-flight requests.

## Error Handling

Errors are wrapped with `%w` so callers can use `errors.Is`. Domain errors include:

- `storage.ErrNoSavedPages`
- `storage.ErrDuplicateLink`
- `service.ErrInvalidURL`

## Shutdown

`app/` uses `signal.NotifyContext`. When `SIGINT` or `SIGTERM` arrives, polling stops, in-flight HTTP requests receive context cancellation, and sqlite is closed. Because processed updates persist the next offset in sqlite, restarts resume from the last acknowledged update instead of replaying the full unconfirmed batch.
