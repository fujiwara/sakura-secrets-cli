package sscli

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	sm "github.com/sacloud/sacloud-sdk-go/api/secretmanager"
	v1 "github.com/sacloud/sacloud-sdk-go/api/secretmanager/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
)

func Run(ctx context.Context) error {
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

func newSMClient() (*v1.Client, error) {
	var sa saclient.Client
	return sm.NewClient(&sa)
}
