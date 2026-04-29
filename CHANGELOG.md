# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-04-29

### Added

- Production-ready env configuration via `BOT_TOKEN`, `DATABASE_PATH`, `SEND_EVERY_N`, `LOG_LEVEL`, and `ENV`.
- `service/` layer for link validation, duplicate handling, per-user counters, and scheduled random-link selection.
- sqlite schema initialization with `links` and `message_counters` tables.
- Structured logging with `log/slog`.
- Context-aware consumer, storage, and Telegram client calls.
- Graceful shutdown on `SIGINT` and `SIGTERM`.
- Unit tests for service behavior, URL validation, and sqlite storage.
- GitHub Actions CI for formatting, vet, tests, and race tests.
- Project documentation, architecture notes, contributing guide, `.env.example`, and Makefile.

### Changed

- Telegram event processing now delegates business rules to the service layer.
- Duplicate links are ignored per user and reported to the user.
- `/rnd` returns a random saved link without removing it from storage.

### Fixed

- Removed hardcoded bot token and database path from runtime wiring.
- Fixed Telegram JSON tags and import casing for the existing `clients/Telegram` package.
