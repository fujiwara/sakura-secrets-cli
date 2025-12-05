# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A CLI tool for managing secrets in SAKURA Cloud's Secret Manager service. Built with Go and uses the `sacloud/secretmanager-api-go` SDK.

## Build Commands

```bash
# Build the binary
make

# Run tests
make test

# Install to $GOPATH/bin
make install

# Build release binaries (requires goreleaser)
make dist

# Clean build artifacts
make clean
```

## Architecture

- **Package**: `sscli` (root package)
- **Entry point**: `cmd/sakura-secrets-cli/main.go`
- **CLI framework**: [Kong](https://github.com/alecthomas/kong)
- **API client**: `github.com/sacloud/secretmanager-api-go`

### File Structure

| File | Purpose |
|------|---------|
| `cli.go` | CLI struct definitions and command flags |
| `main.go` | Command routing via Kong |
| `list.go`, `get.go`, `create.go`, `update.go`, `delete.go` | Individual command implementations |
| `version.go` | Version variable (set at build time) |

### Commands

All commands require `--vault-id` flag or `VAULT_ID` environment variable:

- `list` - List all secrets
- `get <name>` - Get a secret value (supports `--secret-version`)
- `create <name> [value]` - Create a new secret (`--stdin` to read from stdin)
- `update <name> [value]` - Update an existing secret (`--stdin` to read from stdin)
- `delete <name>` - Delete a secret (`--force` to skip confirmation)

### Authentication

The SAKURA Cloud API credentials are handled by the `sacloud/secretmanager-api-go` SDK (typically via environment variables like `SAKURACLOUD_ACCESS_TOKEN` and `SAKURACLOUD_ACCESS_TOKEN_SECRET`).
