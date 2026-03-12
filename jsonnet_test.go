package sscli

import (
	"testing"

	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
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
