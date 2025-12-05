package sscli

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

func Run(ctx context.Context) error {
	c := &CLI{}
	k, err := kong.New(c, kong.Vars{"version": fmt.Sprintf("simplemq-cli %s", Version)})
	if err != nil {
		return fmt.Errorf("failed to create kong: %w", err)
	}
	kx, err := k.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("failed to parse command line: %w", err)
	}
	switch kx.Command() {
	case "list":
		return runListCommand(ctx, c)
	case "get <name>":
		return runGetCommand(ctx, c)
	case "create <name> <value>", "create <name>":
		return runCreateCommand(ctx, c)
	case "update <name> <value>", "update <name>":
		return runUpdateCommand(ctx, c)
	case "delete <name>":
		return runDeleteCommand(ctx, c)
	default:
		return fmt.Errorf("unknown command: %s", kx.Command())
	}
}
