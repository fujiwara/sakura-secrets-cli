package sscli

import (
	"testing"

	jsonnet "github.com/google/go-jsonnet"
	sm "github.com/sacloud/sacloud-sdk-go/api/secretmanager"
	v1 "github.com/sacloud/sacloud-sdk-go/api/secretmanager/apis/v1"

	"github.com/sacloud/sakumock/secretmanager"
)

func TestParseSecretName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantVersion v1.OptNilInt
		wantErr     bool
	}{
		{
			name:     "name only",
			input:    "my-secret",
			wantName: "my-secret",
		},
		{
			name:        "name with version",
			input:       "my-secret:3",
			wantName:    "my-secret",
			wantVersion: v1.NewOptNilInt(3),
		},
		{
			name:    "invalid version",
			input:   "my-secret:abc",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, version, err := parseSecretName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseSecretName(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSecretName(%q) unexpected error: %v", tt.input, err)
			}
			if name != tt.wantName {
				t.Errorf("parseSecretName(%q) name = %q, want %q", tt.input, name, tt.wantName)
			}
			if version != tt.wantVersion {
				t.Errorf("parseSecretName(%q) version = %v, want %v", tt.input, version, tt.wantVersion)
			}
		})
	}
}

func TestSecretNativeFunction(t *testing.T) {
	fn := SecretNativeFunction(t.Context())
	if fn.Name != "secret" {
		t.Errorf("Name = %q, want %q", fn.Name, "secret")
	}
	if len(fn.Params) != 2 {
		t.Fatalf("Params = %v, want 2 params", fn.Params)
	}
	if string(fn.Params[0]) != "vault_id" {
		t.Errorf("Params[0] = %q, want %q", fn.Params[0], "vault_id")
	}
	if string(fn.Params[1]) != "name" {
		t.Errorf("Params[1] = %q, want %q", fn.Params[1], "name")
	}
}

// setupLocalServer starts a local SecretManager mock server and sets
// environment variables so that newSMClient() connects to it.
func setupLocalServer(t *testing.T) *secretmanager.Server {
	t.Helper()
	srv := secretmanager.NewTestServer(secretmanager.Config{})
	t.Cleanup(srv.Close)
	t.Setenv("SAKURA_ENDPOINTS_SECRETMANAGER", srv.TestURL())
	t.Setenv("SAKURA_ACCESS_TOKEN", "dummy")
	t.Setenv("SAKURA_ACCESS_TOKEN_SECRET", "dummy")
	return srv
}

func TestSecretNativeFunctionWithJsonnet(t *testing.T) {
	setupLocalServer(t)
	ctx := t.Context()

	// Create secrets via unveilSecret's underlying client
	createSecret(t, "test-vault", "db-password", "s3cret123")
	createSecret(t, "test-vault", "api-key", "key-v1")
	// Update to create version 2
	createSecret(t, "test-vault", "api-key", "key-v2")

	t.Run("read latest secret", func(t *testing.T) {
		vm := jsonnet.MakeVM()
		vm.NativeFunction(SecretNativeFunction(ctx))
		output, err := vm.EvaluateAnonymousSnippet("test.jsonnet",
			`{ password: std.native("secret")("test-vault", "db-password") }`,
		)
		if err != nil {
			t.Fatal(err)
		}
		want := "{\n   \"password\": \"s3cret123\"\n}\n"
		if output != want {
			t.Errorf("got %q, want %q", output, want)
		}
	})

	t.Run("read specific version", func(t *testing.T) {
		vm := jsonnet.MakeVM()
		vm.NativeFunction(SecretNativeFunction(ctx))
		output, err := vm.EvaluateAnonymousSnippet("test.jsonnet",
			`{ key: std.native("secret")("test-vault", "api-key:1") }`,
		)
		if err != nil {
			t.Fatal(err)
		}
		want := "{\n   \"key\": \"key-v1\"\n}\n"
		if output != want {
			t.Errorf("got %q, want %q", output, want)
		}
	})

	t.Run("read latest version", func(t *testing.T) {
		vm := jsonnet.MakeVM()
		vm.NativeFunction(SecretNativeFunction(ctx))
		output, err := vm.EvaluateAnonymousSnippet("test.jsonnet",
			`{ key: std.native("secret")("test-vault", "api-key") }`,
		)
		if err != nil {
			t.Fatal(err)
		}
		want := "{\n   \"key\": \"key-v2\"\n}\n"
		if output != want {
			t.Errorf("got %q, want %q", output, want)
		}
	})

	t.Run("secret not found", func(t *testing.T) {
		vm := jsonnet.MakeVM()
		vm.NativeFunction(SecretNativeFunction(ctx))
		_, err := vm.EvaluateAnonymousSnippet("test.jsonnet",
			`std.native("secret")("test-vault", "nonexistent")`,
		)
		if err == nil {
			t.Fatal("expected error for non-existent secret")
		}
	})
}

// createSecret creates a secret via the API (reuses the same env-var-based client).
func createSecret(t *testing.T, vaultID, name, value string) {
	t.Helper()
	ctx := t.Context()
	client, err := newSMClient()
	if err != nil {
		t.Fatal(err)
	}
	secOp := sm.NewSecretOp(client, vaultID)
	if _, err := secOp.Create(ctx, v1.CreateSecret{Name: name, Value: value}); err != nil {
		t.Fatal(err)
	}
}
