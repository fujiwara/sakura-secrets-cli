package sscli

import (
	"context"
	"fmt"

	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

func runGetCommand(ctx context.Context, cli *CLI) error {
	cmd := cli.Get
	client, err := sm.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create SecretManager client: %w", err)
	}
	secOp := sm.NewSecretOp(client, cli.VaultID)
	res, err := secOp.Unveil(ctx, v1.Unveil{
		Name:    cmd.Name,
		Version: v1.NewOptNilInt(cmd.SecretVersion),
	})
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}
	fmt.Println(jsonString(res))
	return nil
}
