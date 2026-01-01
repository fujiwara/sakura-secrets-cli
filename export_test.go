package sscli

import (
	"testing"
)

func TestParseNameParam(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantName   string
		wantVer    int
		wantJSON   bool
		wantPrefix string
		wantErr    bool
	}{
		{
			name:       "name only",
			input:      "foo",
			wantName:   "foo",
			wantVer:    0,
			wantJSON:   false,
			wantPrefix: "",
		},
		{
			name:       "name with version",
			input:      "foo:1",
			wantName:   "foo",
			wantVer:    1,
			wantJSON:   false,
			wantPrefix: "",
		},
		{
			name:       "name with empty version",
			input:      "foo:",
			wantName:   "foo",
			wantVer:    0,
			wantJSON:   false,
			wantPrefix: "",
		},
		{
			name:       "name with version and json",
			input:      "foo:2:json",
			wantName:   "foo",
			wantVer:    2,
			wantJSON:   true,
			wantPrefix: "",
		},
		{
			name:       "name with empty version and json",
			input:      "foo::json",
			wantName:   "foo",
			wantVer:    0,
			wantJSON:   true,
			wantPrefix: "",
		},
		{
			name:       "name with version, json and prefix",
			input:      "foo:3:json:DB_",
			wantName:   "foo",
			wantVer:    3,
			wantJSON:   true,
			wantPrefix: "DB_",
		},
		{
			name:       "name with empty version, json and prefix",
			input:      "foo::json:MYAPP_",
			wantName:   "foo",
			wantVer:    0,
			wantJSON:   true,
			wantPrefix: "MYAPP_",
		},
		{
			name:       "non-json third part",
			input:      "foo:1:notjson",
			wantName:   "foo",
			wantVer:    1,
			wantJSON:   false,
			wantPrefix: "",
		},
		{
			name:    "invalid version",
			input:   "foo:abc",
			wantErr: true,
		},
		{
			name:    "invalid version with json",
			input:   "foo:abc:json",
			wantErr: true,
		},
		{
			name:    "too many colons",
			input:   "foo:1:json:prefix:extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ver, isJSON, prefix, err := parseNameParam(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseNameParam(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseNameParam(%q) unexpected error: %v", tt.input, err)
				return
			}
			if name != tt.wantName {
				t.Errorf("parseNameParam(%q) name = %q, want %q", tt.input, name, tt.wantName)
			}
			if ver != tt.wantVer {
				t.Errorf("parseNameParam(%q) version = %d, want %d", tt.input, ver, tt.wantVer)
			}
			if isJSON != tt.wantJSON {
				t.Errorf("parseNameParam(%q) isJSON = %v, want %v", tt.input, isJSON, tt.wantJSON)
			}
			if prefix != tt.wantPrefix {
				t.Errorf("parseNameParam(%q) prefix = %q, want %q", tt.input, prefix, tt.wantPrefix)
			}
		})
	}
}
