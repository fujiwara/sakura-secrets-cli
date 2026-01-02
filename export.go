package sscli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

type ExportCommand struct {
	Name     []string `help:"Names of the secrets to export. You can specify version and options like 'name:version:json:prefix'." required:""`
	Commands []string `arg:"" help:"Command to run with exported secrets in environment variables" optional:""`
}

func ExportEnvs(ctx context.Context, vaultID string, names []string) (map[string]string, error) {
	client, err := sm.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create SecretManager client: %w", err)
	}
	secOp := sm.NewSecretOp(client, vaultID)
	envs := make(map[string]string)
	for _, np := range names {
		name, version, isJSON, prefix, err := parseNameParam(np)
		if err != nil {
			return nil, err
		}
		res, err := secOp.Unveil(ctx, v1.Unveil{
			Name:    name,
			Version: v1.NewOptNilInt(version),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get secret: %w", err)
		}
		if isJSON {
			var m map[string]string
			if err := json.Unmarshal([]byte(res.Value), &m); err != nil {
				return nil, fmt.Errorf("failed to parse secret value as JSON object: %w", err)
			}
			for k, v := range m {
				envs[makeExportEnvKey(k, prefix)] = v
			}
			continue
		} else {
			envs[makeExportEnvKey(name, prefix)] = res.Value
		}
	}
	return envs, nil
}

var EnvKeyInvalidRegex = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// makeExportEnvKey formats the environment variable key by replacing invalid characters
// with underscores and converting it to uppercase, with an optional prefix.
func makeExportEnvKey(name, prefix string) string {
	return strings.ToUpper(EnvKeyInvalidRegex.ReplaceAllString(prefix+name, "_"))
}

func runExportCommand(ctx context.Context, cli *CLI) error {
	cmd := cli.Secret.Export
	envMap, err := ExportEnvs(ctx, cli.Secret.VaultID, cmd.Name)
	if err != nil {
		return err
	}
	if len(cmd.Commands) > 0 {
		envs := make([]string, 0, len(envMap))
		for k, v := range envMap {
			envs = append(envs, fmt.Sprintf("%s=%s", k, v))
		}
		return runCommandWithEnvs(ctx, cli, envs, cmd.Commands)
	}
	for k, v := range envMap {
		fmt.Printf("export %s=%s\n", k, v)
	}
	return nil
}

func runCommandWithEnvs(ctx context.Context, cli *CLI, envs []string, command []string) error {
	bin, err := exec.LookPath(command[0])
	if err != nil {
		return fmt.Errorf("command is not executable %s: %w", command[0], err)
	}

	return syscall.Exec(bin, command, append(os.Environ(), envs...))
}

func parseVersionString(s string) (int, error) {
	if len(s) == 0 {
		return 0, nil
	} else {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid version '%s': %w", s, err)
		}
		return int(v), nil
	}
}

func parseNameParam(nameParam string) (string, int, bool, string, error) {
	switch strings.Count(nameParam, ":") {
	case 0:
		return nameParam, 0, false, "", nil
	case 1:
		parts := strings.SplitN(nameParam, ":", 2)
		name := parts[0]
		version, err := parseVersionString(parts[1])
		return name, version, false, "", err
	case 2: // for future extension like name:version:json
		parts := strings.SplitN(nameParam, ":", 3)
		name := parts[0]
		version, err := parseVersionString(parts[1])
		if err != nil {
			return "", 0, false, "", err
		}
		return name, version, parts[2] == "json", "", nil
	case 3: // for future extension like name:version:json:prefix
		parts := strings.SplitN(nameParam, ":", 4)
		name := parts[0]
		version, err := parseVersionString(parts[1])
		if err != nil {
			return "", 0, false, "", err
		}
		return name, version, parts[2] == "json", parts[3], nil
	default:
		return "", 0, false, "", fmt.Errorf("invalid name parameter: %s", nameParam)
	}
}
