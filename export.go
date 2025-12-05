package sscli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

type ExportCommand struct {
	Names    []string `arg:"" help:"Names of the secrets to export e.g. foo:1 for version 1, foo for latest version"`
	FromJSON bool     `help:"parse value as JSON object and export each key as separate secret"`
}

func runExportCommand(ctx context.Context, cli *CLI) error {
	cmd := cli.Secret.Export
	client, err := sm.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create SecretManager client: %w", err)
	}
	secOp := sm.NewSecretOp(client, cli.Secret.VaultID)
	for _, nameWithVersion := range cmd.Names {
		var name string
		var version int
		strings.SplitN(nameWithVersion, ":", 2)
		parts := strings.SplitN(nameWithVersion, ":", 2)
		name = parts[0]
		if len(parts) == 2 {
			v, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid version in '%s': %w", nameWithVersion, err)
			}
			version = int(v)
		} else {
			version = 0
		}
		res, err := secOp.Unveil(ctx, v1.Unveil{
			Name:    name,
			Version: v1.NewOptNilInt(version),
		})
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}
		if cmd.FromJSON {
			var m map[string]string
			if err := json.Unmarshal([]byte(res.Value), &m); err != nil {
				return fmt.Errorf("failed to parse secret value as JSON object: %w", err)
			}
			for k, v := range m {
				fmt.Printf("export %s=%s\n", k, v)
			}
			continue
		} else {
			fmt.Printf("export %s=%s\n", name, res.Value)
		}
	}
	return nil
}
