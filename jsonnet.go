package sscli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

// SecretNativeFunction returns a Jsonnet native function that reads secrets
// from SAKURA Cloud Secret Manager.
//
// Usage in Jsonnet: secret("vault-id", "name") or secret("vault-id", "name:version")
func SecretNativeFunction(ctx context.Context) *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   "secret",
		Params: ast.Identifiers{"vault_id", "name"},
		Func: func(args []any) (any, error) {
			vaultID, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("secret: vault_id must be a string")
			}
			name, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("secret: name must be a string")
			}
			return unveilSecret(ctx, vaultID, name)
		},
	}
}

// parseSecretName parses a name string into a secret name and optional version.
// Format: "name" (latest) or "name:version" (specific version).
func parseSecretName(s string) (string, v1.OptNilInt, error) {
	name, versionStr, hasVersion := strings.Cut(s, ":")
	if !hasVersion {
		return name, v1.OptNilInt{}, nil
	}
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return "", v1.OptNilInt{}, fmt.Errorf("secret: invalid version %q: %w", versionStr, err)
	}
	return name, v1.NewOptNilInt(version), nil
}

func unveilSecret(ctx context.Context, vaultID, nameWithVersion string) (string, error) {
	name, version, err := parseSecretName(nameWithVersion)
	if err != nil {
		return "", err
	}
	client, err := newSMClient()
	if err != nil {
		return "", fmt.Errorf("secret: failed to create client: %w", err)
	}
	secOp := sm.NewSecretOp(client, vaultID)
	result, err := secOp.Unveil(ctx, v1.Unveil{Name: name, Version: version})
	if err != nil {
		return "", fmt.Errorf("secret: failed to unveil %q: %w", nameWithVersion, err)
	}
	return result.Value, nil
}
