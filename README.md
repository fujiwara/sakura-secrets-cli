# sakura-secrets-cli

A CLI tool for managing secrets in [SAKURA Cloud Secret Manager](https://manual.sakura.ad.jp/cloud/appliance/secret-manager/).

## Installation

```bash
go install github.com/fujiwara/sakura-secrets-cli/cmd/sakura-secrets-cli@latest
```

Or download the binary from [Releases](https://github.com/fujiwara/sakura-secrets-cli/releases).

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `SAKURACLOUD_ACCESS_TOKEN` | SAKURA Cloud API access token | Yes |
| `SAKURACLOUD_ACCESS_TOKEN_SECRET` | SAKURA Cloud API access token secret | Yes |
| `VAULT_ID` | Secret Manager vault ID | Yes (or use `--vault-id` flag) |

## Usage

```
Usage: sakura-secrets-cli --vault-id=STRING <command> [flags]

Flags:
  -h, --help               Show context-sensitive help.
      --vault-id=STRING    Vault ID ($VAULT_ID)
  -v, --version            Show version and exit.

Commands:
  list --vault-id=STRING [flags]
    List secrets

  get --vault-id=STRING <name> [flags]
    Get secret value

  create --vault-id=STRING <name> [<value>] [flags]
    Create a new secret

  update --vault-id=STRING <name> [<value>] [flags]
    Update an existing secret

  delete --vault-id=STRING <name> [flags]
    Delete a secret

Run "sakura-secrets-cli <command> --help" for more information on a command.
```

### Commands

#### List secrets

```bash
sakura-secrets-cli list
```

#### Get a secret

```bash
sakura-secrets-cli get <name>

# Get a specific version
sakura-secrets-cli get <name> --secret-version 1
```

#### Create a secret

```bash
sakura-secrets-cli create <name> <value>

# Read value from stdin
echo "secret-value" | sakura-secrets-cli create <name> --stdin
```

#### Update a secret

```bash
sakura-secrets-cli update <name> <value>

# Read value from stdin
cat secret.txt | sakura-secrets-cli update <name> --stdin
```

#### Delete a secret

```bash
sakura-secrets-cli delete <name>

# Skip confirmation prompt
sakura-secrets-cli delete <name> --force
```

## LICENSE

MIT

## Author

fujiwara
