package localserver_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fujiwara/sakura-secrets-cli/localserver"
)

const testPrefix = "/api/cloud/1.1"
const testVaultID = "test-vault-123"

func secretsURL(base string) string {
	return base + testPrefix + "/secretmanager/vaults/" + testVaultID + "/secrets"
}

func unveilURL(base string) string {
	return secretsURL(base) + "/unveil"
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func mustReadJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("failed to parse JSON: %s, body: %s", err, string(body))
	}
}

func TestSecretLifecycle(t *testing.T) {
	srv := httptest.NewServer(localserver.NewServer(testPrefix))
	defer srv.Close()

	// List: initially empty
	resp, err := http.Get(secretsURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var listResp struct {
		Count   int `json:"Count"`
		Secrets []struct {
			Name          string `json:"Name"`
			LatestVersion int    `json:"LatestVersion"`
		} `json:"Secrets"`
	}
	mustReadJSON(t, resp, &listResp)
	if listResp.Count != 0 {
		t.Fatalf("expected 0 secrets, got %d", listResp.Count)
	}

	// Create secret "foo"
	createBody := mustMarshal(t, map[string]any{
		"Secret": map[string]any{"Name": "foo", "Value": "bar"},
	})
	resp, err = http.Post(secretsURL(srv.URL), "application/json", bytes.NewReader(createBody))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var createResp struct {
		Secret struct {
			Name          string `json:"Name"`
			LatestVersion int    `json:"LatestVersion"`
		} `json:"Secret"`
	}
	mustReadJSON(t, resp, &createResp)
	if createResp.Secret.Name != "foo" || createResp.Secret.LatestVersion != 1 {
		t.Fatalf("unexpected create response: %+v", createResp)
	}

	// List: 1 secret
	resp, err = http.Get(secretsURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	mustReadJSON(t, resp, &listResp)
	if listResp.Count != 1 {
		t.Fatalf("expected 1 secret, got %d", listResp.Count)
	}
	if listResp.Secrets[0].Name != "foo" || listResp.Secrets[0].LatestVersion != 1 {
		t.Fatalf("unexpected list item: %+v", listResp.Secrets[0])
	}

	// Unveil secret "foo" (latest)
	unveilBody := mustMarshal(t, map[string]any{
		"Secret": map[string]any{"Name": "foo", "Version": nil},
	})
	resp, err = http.Post(unveilURL(srv.URL), "application/json", bytes.NewReader(unveilBody))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var unveilResp struct {
		Secret struct {
			Name    string `json:"Name"`
			Version int    `json:"Version"`
			Value   string `json:"Value"`
		} `json:"Secret"`
	}
	mustReadJSON(t, resp, &unveilResp)
	if unveilResp.Secret.Name != "foo" || unveilResp.Secret.Version != 1 || unveilResp.Secret.Value != "bar" {
		t.Fatalf("unexpected unveil response: %+v", unveilResp)
	}

	// Update secret "foo" (create v2)
	updateBody := mustMarshal(t, map[string]any{
		"Secret": map[string]any{"Name": "foo", "Value": "baz"},
	})
	resp, err = http.Post(secretsURL(srv.URL), "application/json", bytes.NewReader(updateBody))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	mustReadJSON(t, resp, &createResp)
	if createResp.Secret.LatestVersion != 2 {
		t.Fatalf("expected version 2, got %d", createResp.Secret.LatestVersion)
	}

	// Unveil version 1
	unveilV1Body := mustMarshal(t, map[string]any{
		"Secret": map[string]any{"Name": "foo", "Version": 1},
	})
	resp, err = http.Post(unveilURL(srv.URL), "application/json", bytes.NewReader(unveilV1Body))
	if err != nil {
		t.Fatal(err)
	}
	mustReadJSON(t, resp, &unveilResp)
	if unveilResp.Secret.Value != "bar" || unveilResp.Secret.Version != 1 {
		t.Fatalf("expected v1 value 'bar', got: %+v", unveilResp.Secret)
	}

	// Unveil latest (should be v2)
	resp, err = http.Post(unveilURL(srv.URL), "application/json", bytes.NewReader(unveilBody))
	if err != nil {
		t.Fatal(err)
	}
	mustReadJSON(t, resp, &unveilResp)
	if unveilResp.Secret.Value != "baz" || unveilResp.Secret.Version != 2 {
		t.Fatalf("expected v2 value 'baz', got: %+v", unveilResp.Secret)
	}

	// Delete secret "foo"
	deleteBody := mustMarshal(t, map[string]any{
		"Secret": map[string]any{"Name": "foo"},
	})
	req, err := http.NewRequest(http.MethodDelete, secretsURL(srv.URL), bytes.NewReader(deleteBody))
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	// List: empty again
	resp, err = http.Get(secretsURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	mustReadJSON(t, resp, &listResp)
	if listResp.Count != 0 {
		t.Fatalf("expected 0 secrets after delete, got %d", listResp.Count)
	}
}

func TestUnveilNotFound(t *testing.T) {
	srv := httptest.NewServer(localserver.NewServer(testPrefix))
	defer srv.Close()

	body := mustMarshal(t, map[string]any{
		"Secret": map[string]any{"Name": "nonexistent", "Version": nil},
	})
	resp, err := http.Post(unveilURL(srv.URL), "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteNotFound(t *testing.T) {
	srv := httptest.NewServer(localserver.NewServer(testPrefix))
	defer srv.Close()

	body := mustMarshal(t, map[string]any{
		"Secret": map[string]any{"Name": "nonexistent"},
	})
	req, err := http.NewRequest(http.MethodDelete, secretsURL(srv.URL), bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestMultipleSecrets(t *testing.T) {
	srv := httptest.NewServer(localserver.NewServer(testPrefix))
	defer srv.Close()

	for _, name := range []string{"alpha", "beta", "gamma"} {
		body := mustMarshal(t, map[string]any{
			"Secret": map[string]any{"Name": name, "Value": "value-" + name},
		})
		resp, err := http.Post(secretsURL(srv.URL), "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	}

	resp, err := http.Get(secretsURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	var listResp struct {
		Count   int `json:"Count"`
		Secrets []struct {
			Name string `json:"Name"`
		} `json:"Secrets"`
	}
	mustReadJSON(t, resp, &listResp)
	if listResp.Count != 3 {
		t.Fatalf("expected 3 secrets, got %d", listResp.Count)
	}
	// Sorted alphabetically
	if listResp.Secrets[0].Name != "alpha" || listResp.Secrets[1].Name != "beta" || listResp.Secrets[2].Name != "gamma" {
		t.Fatalf("unexpected order: %+v", listResp.Secrets)
	}
}

func TestDifferentVaults(t *testing.T) {
	srv := httptest.NewServer(localserver.NewServer(testPrefix))
	defer srv.Close()

	vault1URL := srv.URL + testPrefix + "/secretmanager/vaults/vault-1/secrets"
	vault2URL := srv.URL + testPrefix + "/secretmanager/vaults/vault-2/secrets"

	body := mustMarshal(t, map[string]any{
		"Secret": map[string]any{"Name": "secret1", "Value": "value1"},
	})
	resp, err := http.Post(vault1URL, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// vault-2 should be empty
	resp, err = http.Get(vault2URL)
	if err != nil {
		t.Fatal(err)
	}
	var listResp struct {
		Count int `json:"Count"`
	}
	mustReadJSON(t, resp, &listResp)
	if listResp.Count != 0 {
		t.Fatalf("vault-2 should be empty, got %d", listResp.Count)
	}

	// vault-1 should have 1
	resp, err = http.Get(vault1URL)
	if err != nil {
		t.Fatal(err)
	}
	mustReadJSON(t, resp, &listResp)
	if listResp.Count != 1 {
		t.Fatalf("vault-1 should have 1 secret, got %d", listResp.Count)
	}
}
