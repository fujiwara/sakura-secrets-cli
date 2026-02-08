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
| `SAKURA_ACCESS_TOKEN` | SAKURA Cloud API access token | Yes |
| `SAKURA_ACCESS_TOKEN_SECRET` | SAKURA Cloud API access token secret | Yes |
| `VAULT_ID` | Secret Manager vault ID | Yes (or use `--vault-id` flag) |

`SAKURACLOUD_ACCESS_TOKEN` / `SAKURACLOUD_ACCESS_TOKEN_SECRET` are also supported for backward compatibility.

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

# Get a specific version (two ways)
$ sakura-secrets-cli secret get foo --secret-version 1
{"Name":"foo","Version":1,"Value":"FOO_VALUE"}

$ sakura-secrets-cli secret get foo:1
{"Name":"foo","Version":1,"Value":"FOO_VALUE"}

# Output only the value
$ sakura-secrets-cli secret get foo --value-only
FOO_VALUE
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

The `--name` flag accepts a flexible format: `name[:version][:json][:prefix]`

| Format | Description |
|--------|-------------|
| `name` | Export the latest version |
| `name:1` | Export version 1 |
| `name::json` | Parse JSON and export each key (latest version) |
| `name:1:json` | Parse JSON from version 1 |
| `name::json:DB_` | Parse JSON with prefix `DB_` |
| `name:1:json:DB_` | Parse JSON from version 1 with prefix `DB_` |

**Note:** Environment variable names are automatically converted to uppercase.

##### Output export statements

```bash
$ sakura-secrets-cli secret export --name foo
export FOO=FOO_VALUE

$ sakura-secrets-cli secret export --name foo --name bar
export FOO=FOO_VALUE
export BAR=BAR_VALUE

# Specific version
$ sakura-secrets-cli secret export --name foo:1
export FOO=FOO_VALUE
```

##### Parse JSON secrets

If a secret value is a JSON object, use the `:json` option to expand each key as a separate environment variable:

```bash
$ sakura-secrets-cli secret get jsonvalue
{"Name":"jsonvalue","Version":1,"Value":"{\"db_host\":\"localhost\",\"db_password\":\"secret\"}"}

$ sakura-secrets-cli secret export --name jsonvalue::json
export DB_HOST=localhost
export DB_PASSWORD=secret

# With prefix
$ sakura-secrets-cli secret export --name jsonvalue::json:MYAPP_
export MYAPP_DB_HOST=localhost
export MYAPP_DB_PASSWORD=secret
```

##### Run commands with secrets injected

Run any command with secrets as environment variables. The command receives secrets without any code changes:

```bash
# Run application with secrets
$ sakura-secrets-cli secret export --name db_host --name db_password -- ./my-app

# Expand JSON secrets and run command
$ sakura-secrets-cli secret export --name db_credentials::json -- psql

# Verify with env
$ sakura-secrets-cli secret export --name foo -- env | grep FOO
FOO=FOO_VALUE
```

##### Use with eval for current shell

```bash
$ eval $(sakura-secrets-cli secret export --name api_key)
$ echo $API_KEY
```

## Go Library Usage

You can use this package as a Go library to fetch secrets programmatically.

```go
package main

import (
	"context"
	"fmt"
	"log"

	sscli "github.com/fujiwara/sakura-secrets-cli"
)

func main() {
	ctx := context.Background()
	vaultID := "your-vault-id"

	// Fetch secrets as a map of environment variables
	// Name format: name[:version][:json][:prefix]
	envs, err := sscli.ExportEnvs(ctx, vaultID, []string{
		"db_password",        // latest version
		"api_key:1",          // specific version
		"config::json",       // parse JSON
		"app::json:MYAPP_",   // parse JSON with prefix
	})
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range envs {
		fmt.Printf("%s=%s\n", k, v)
	}
}
```

**Note:** Requires `SAKURA_ACCESS_TOKEN` and `SAKURA_ACCESS_TOKEN_SECRET` environment variables (`SAKURACLOUD_*` variants are also supported).

## LICENSE

MIT

## Author

fujiwara
