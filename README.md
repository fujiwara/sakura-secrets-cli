# sakura-secrets-cli

A CLI tool for managing secrets in [SAKURA Cloud Secret Manager](https://cloud.sakura.ad.jp/products/secrets-manager/).

## Features

- List, get, create, update, and delete secrets
- Export secrets as environment variables for shell scripts
- Run commands with secrets injected as environment variables
- Support for versioned secrets
- Parse JSON-formatted secrets into individual environment variables

## Use Cases

- **Zero-touch deployment**: `sakura-secrets-cli secret export --name DB_PASSWORD -- ./my-app` - your app doesn't need to know about secrets management
- **CI/CD pipelines**: Inject secrets into build and deployment processes without storing them in configuration files
- **Local development**: Run applications locally with production-like secrets management
- **Container environments**: Pass secrets to containers at runtime without baking them into images
- **Shell scripts**: Source exported secrets directly in shell scripts with `eval $(sakura-secrets-cli secret export --name ...)`

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

  secret export --vault-id=STRING [<commands> ...] [flags]
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

##### Output export statements

```bash
$ sakura-secrets-cli secret export --name foo
export foo=FOO_VALUE

$ sakura-secrets-cli secret export --name foo --name bar
export foo=FOO_VALUE
export bar=BAR_VALUE

# Specific version
$ sakura-secrets-cli secret export --name foo:1
export foo=FOO_VALUE
```

##### Parse JSON secrets with `--from-json`

If a secret value is a JSON object, `--from-json` expands each key as a separate environment variable:

```bash
$ sakura-secrets-cli secret get jsonvalue
{"Name":"jsonvalue","Version":1,"Value":"{\"DB_HOST\":\"localhost\",\"DB_PASSWORD\":\"secret\"}"}

$ sakura-secrets-cli secret export --name jsonvalue --from-json
export DB_HOST=localhost
export DB_PASSWORD=secret
```

##### Run commands with secrets injected

Run any command with secrets as environment variables. The command receives secrets without any code changes:

```bash
# Run application with secrets
$ sakura-secrets-cli secret export --name DB_HOST --name DB_PASSWORD -- ./my-app

# Expand JSON secrets and run command
$ sakura-secrets-cli secret export --name db_credentials --from-json -- psql

# Verify with env
$ sakura-secrets-cli secret export --name foo -- env | grep foo
foo=FOO_VALUE
```

##### Use with eval for current shell

```bash
$ eval $(sakura-secrets-cli secret export --name API_KEY)
$ echo $API_KEY
```

## LICENSE

MIT

## Author

fujiwara
