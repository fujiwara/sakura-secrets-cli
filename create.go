package sscli

import (
	"context"
	"fmt"
	"io"
	"os"

	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

type CreateCommand struct {
	Name  string `arg:"" help:"Name of the secret to create"`
	Value string `arg:"" help:"Value of the secret to create" optional:""`
	Stdin bool   `help:"Read value from stdin instead of argument"`
}

func runCreateCommand(ctx context.Context, cli *CLI) error {
	cmd := cli.Secret.Create
	client, err := newSMClient()
	if err != nil {
		return fmt.Errorf("failed to create SecretManager client: %w", err)
	}
	value := cmd.Value
	if cmd.Stdin {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		value = string(b)
	}

	secOp := sm.NewSecretOp(client, cli.Secret.VaultID)
	res, err := secOp.Create(ctx, v1.CreateSecret{
		Name:  cmd.Name,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}
	fmt.Println(jsonString(res))
	return nil
}
