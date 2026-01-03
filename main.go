package sscli

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

func Run(ctx context.Context) error {
	setEnvForCompatibility()
	c := &CLI{}
	k, err := kong.New(c, kong.Vars{"version": fmt.Sprintf("sakura-secrets-cli %s", Version)})
	if err != nil {
		return fmt.Errorf("failed to create kong: %w", err)
	}
	kx, err := k.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("failed to parse command line: %w", err)
	}
	switch kx.Command() {
	case "secret list":
		return runListCommand(ctx, c)
	case "secret get <name>":
		return runGetCommand(ctx, c)
	case "secret create <name> <value>", "secret create <name>":
		return runCreateCommand(ctx, c)
	case "secret update <name> <value>", "secret update <name>":
		return runUpdateCommand(ctx, c)
	case "secret delete <name>":
		return runDeleteCommand(ctx, c)
	case "secret export", "secret export <commands>":
		return runExportCommand(ctx, c)
	default:
		return fmt.Errorf("unknown command: %s", kx.Command())
	}
}

func setEnvForCompatibility() {
	// For compatibility with older versions and SDKs
	if v, ok := os.LookupEnv("SAKURA_ACCESS_TOKEN"); ok {
		os.Setenv("SAKURACLOUD_ACCESS_TOKEN", v)
	}
	if v, ok := os.LookupEnv("SAKURA_ACCESS_TOKEN_SECRET"); ok {
		os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", v)
	}
}
