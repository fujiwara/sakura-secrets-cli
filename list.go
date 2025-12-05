package sscli

import (
	"context"
	"encoding/json"
	"fmt"

	sm "github.com/sacloud/secretmanager-api-go"
)

func runListCommand(ctx context.Context, cli *CLI) error {
	client, err := sm.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create SecretManager client: %w", err)
	}
	secOp := sm.NewSecretOp(client, cli.VaultID)
	res, err := secOp.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}
	for _, s := range res {
		fmt.Println(jsonString(s))
	}
	return nil
}

func jsonString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
