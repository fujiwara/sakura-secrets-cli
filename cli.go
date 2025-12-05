package sscli

import "github.com/alecthomas/kong"

type CLI struct {
	VaultID string `help:"Vault ID" required:"" env:"VAULT_ID"`

	List   ListCommand   `cmd:"" help:"List secrets"`
	Get    GetCommand    `cmd:"" help:"Get secret value"`
	Create CreateCommand `cmd:"" help:"Create a new secret"`
	Update UpdateCommand `cmd:"" help:"Update an existing secret"`
	Delete DeleteCommand `cmd:"" help:"Delete a secret"`

	Version kong.VersionFlag `short:"v" help:"Show version and exit."`
}

type ListCommand struct{}

type GetCommand struct {
	Name          string `arg:"" help:"Name of the secret to get"`
	SecretVersion int    `help:"Version of the secret to get" default:"0"`
}

type CreateCommand struct {
	Name  string `arg:"" help:"Name of the secret to create"`
	Value string `arg:"" help:"Value of the secret to create" optional:""`
	Stdin bool   `help:"Read value from stdin instead of argument"`
}

type UpdateCommand struct {
	Name  string `arg:"" help:"Name of the secret to update"`
	Value string `arg:"" help:"New value of the secret" optional:""`
	Stdin bool   `help:"Read value from stdin instead of argument"`
}

type DeleteCommand struct {
	Name  string `arg:"" help:"Name of the secret to delete"`
	Force bool   `help:"Force delete without confirmation"`
}
