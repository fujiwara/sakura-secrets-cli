package sscli

import "github.com/alecthomas/kong"

type CLI struct {
	Secret struct {
		List   ListCommand   `cmd:"" help:"List secrets"`
		Get    GetCommand    `cmd:"" help:"Get secret value"`
		Create CreateCommand `cmd:"" help:"Create a new secret"`
		Update UpdateCommand `cmd:"" help:"Update an existing secret"`
		Delete DeleteCommand `cmd:"" help:"Delete a secret"`
		Export ExportCommand `cmd:"" help:"Export secrets as environment variables"`

		VaultID string `help:"Vault ID" required:"" env:"VAULT_ID"`
	} `cmd:"" help:"Manage secrets in Sakura Secret Manager"`

	Version kong.VersionFlag `short:"v" help:"Show version and exit."`
}
