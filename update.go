package sscli

import (
	"context"
	"fmt"
	"io"
	"os"

	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

type UpdateCommand struct {
	Name  string `arg:"" help:"Name of the secret to update"`
	Value string `arg:"" help:"New value of the secret" optional:""`
	Stdin bool   `help:"Read value from stdin instead of argument"`
}

func runUpdateCommand(ctx context.Context, cli *CLI) error {
	cmd := cli.Secret.Update
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
	res, err := secOp.Update(ctx, v1.CreateSecret{
		Name:  cmd.Name,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}
	fmt.Println(jsonString(res))
	return nil
}
