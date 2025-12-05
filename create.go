package sscli

import (
	"context"
	"fmt"
	"io"
	"os"

	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

func runCreateCommand(ctx context.Context, cli *CLI) error {
	cmd := cli.Create
	client, err := sm.NewClient()
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

	secOp := sm.NewSecretOp(client, cli.VaultID)
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
