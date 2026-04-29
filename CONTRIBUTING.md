# Contributing

## Development Setup

1. Install Go matching `go.mod`.
2. Install a C compiler because sqlite uses `github.com/mattn/go-sqlite3`.
3. Copy `.env.example` to `.env` and fill in `BOT_TOKEN`.
4. Run tests before opening a pull request.

## Commands

```bash
go test ./...
go test -race ./...
go vet ./...
test -z "$(gofmt -l .)"
```

Or use:

```bash
make test
make lint
```

## Code Guidelines

- Keep the existing package layout.
- Put business rules in `service/`.
- Keep `events/telegram` focused on parsing and routing.
- Keep `consumer/` focused on polling orchestration.
- Keep `storage/` focused on data access.
- Wrap errors with `%w` when returning lower-level failures.
- Pass `context.Context` through IO and storage operations.
- Add focused tests for service and storage behavior when changing logic.

## Pull Requests

Before submitting:

- Update documentation when behavior or configuration changes.
- Add or update tests for user-visible behavior.
- Run `make lint` and `make test`.
- Keep unrelated refactors out of production fixes.
