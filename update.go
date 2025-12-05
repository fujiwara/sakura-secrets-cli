package sscli

import (
	"context"
	"fmt"
	"io"
	"os"

	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

func runUpdateCommand(ctx context.Context, cli *CLI) error {
	cmd := cli.Update
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
