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
Usage: sakura-secrets-cli <command> [flags]

Flags:
  -h, --help       Show context-sensitive help.
  -v, --version    Show version and exit.

Commands:
  secret list --vault-id=STRING
    List secrets

  secret get --vault-id=STRING <name> [flags]
    Get secret value

  secret create --vault-id=STRING <name> [<value>] [flags]
    Create a new secret

  secret update --vault-id=STRING <name> [<value>] [flags]
    Update an existing secret

  secret delete --vault-id=STRING <name> [flags]
    Delete a secret

  secret export --vault-id=STRING <names> ... [flags]
    Export secrets as environment variables

Run "sakura-secrets-cli <command> --help" for more information on a command.
```

### Commands

#### List secrets

```bash
$ sakura-secrets-cli secret list
{"Name":"foo","LatestVersion":2}
{"Name":"bar","LatestVersion":2}
{"Name":"jsonvalue","LatestVersion":1}
```

#### Get a secret

```bash
$ sakura-secrets-cli secret get foo
{"Name":"foo","Version":2,"Value":"FOO_VALUE"}

# Get a specific version
$ sakura-secrets-cli secret get foo --secret-version 1
{"Name":"foo","Version":1,"Value":"FOO_VALUE"}
```

#### Create a secret

```bash
$ sakura-secrets-cli secret create my-secret "secret-value"
{"Name":"my-secret","LatestVersion":1}

# Read value from stdin
$ echo "secret-value" | sakura-secrets-cli secret create my-secret --stdin
{"Name":"my-secret","LatestVersion":1}
```

#### Update a secret

```bash
$ sakura-secrets-cli secret update my-secret "new-value"
{"Name":"my-secret","LatestVersion":2}

# Read value from stdin
$ cat secret.txt | sakura-secrets-cli secret update my-secret --stdin
{"Name":"my-secret","LatestVersion":3}
```

#### Delete a secret

```bash
$ sakura-secrets-cli secret delete my-secret
Are you sure you want to delete secret "my-secret"? (yes/no): yes
# (no output on success)

# Skip confirmation prompt
$ sakura-secrets-cli secret delete my-secret --force
# (no output on success)
```

#### Export secrets as environment variables

```bash
# Export a single secret
$ sakura-secrets-cli secret export foo
export foo=FOO_VALUE

# Export multiple secrets
$ sakura-secrets-cli secret export foo bar
export foo=FOO_VALUE
export bar=BAR_VALUE

# Export a specific version
$ sakura-secrets-cli secret export foo:1
export foo=FOO_VALUE
# Parse JSON value and export each key as separate variable
$ sakura-secrets-cli secret get jsonvalue
{"Name":"jsonvalue","Version":1,"Value":"{\"A\":\"AAA\",\"bbbb\":\"BBB\"}"}

$ sakura-secrets-cli secret export jsonvalue --from-json
export A=AAA
export bbbb=BBB
```

## LICENSE

MIT

## Author

fujiwara
