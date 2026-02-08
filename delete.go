package sscli

import (
	"context"
	"fmt"

	"github.com/Songmu/prompter"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

type DeleteCommand struct {
	Name  string `arg:"" help:"Name of the secret to delete"`
	Force bool   `help:"Force delete without confirmation"`
}

func runDeleteCommand(ctx context.Context, cli *CLI) error {
	cmd := cli.Secret.Delete

	if cmd.Force || prompter.YesNo(fmt.Sprintf("Are you sure you want to delete the secret '%s'?", cmd.Name), false) {
		// proceed
	} else {
		fmt.Println("Aborted")
		return nil
	}

	client, err := newSMClient()
	if err != nil {
		return fmt.Errorf("failed to create SecretManager client: %w", err)
	}
	secOp := sm.NewSecretOp(client, cli.Secret.VaultID)
	if err := secOp.Delete(ctx, v1.DeleteSecret{
		Name: cmd.Name,
	}); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}
