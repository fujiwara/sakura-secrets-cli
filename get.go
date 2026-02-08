package sscli

import (
	"context"
	"fmt"

	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

type GetCommand struct {
	Name          string `arg:"" help:"Name of the secret to get"`
	SecretVersion int    `help:"Version of the secret to get" default:"0"`
	ValueOnly     bool   `help:"Output only the value of the secret"`
}

func runGetCommand(ctx context.Context, cli *CLI) error {
	cmd := cli.Secret.Get
	client, err := newSMClient()
	if err != nil {
		return fmt.Errorf("failed to create SecretManager client: %w", err)
	}

	var name string
	var version int
	if cmd.SecretVersion == 0 {
		var err error
		name, version, _, _, err = parseNameParam(cmd.Name)
		if err != nil {
			return err
		}
	} else {
		name = cmd.Name
		version = cmd.SecretVersion
	}

	secOp := sm.NewSecretOp(client, cli.Secret.VaultID)
	res, err := secOp.Unveil(ctx, v1.Unveil{
		Name:    name,
		Version: v1.NewOptNilInt(version),
	})
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}
	if cmd.ValueOnly {
		fmt.Println(res.Value)
	} else {
		fmt.Println(jsonString(res))
	}
	return nil
}
