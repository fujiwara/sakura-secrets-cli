package localserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// JSON request/response types matching the OpenAPI spec.

type secretResponse struct {
	Name          string `json:"Name"`
	LatestVersion int    `json:"LatestVersion"`
}

type wrappedSecret struct {
	Secret secretResponse `json:"Secret"`
}

type createSecretRequest struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type wrappedCreateSecret struct {
	Secret createSecretRequest `json:"Secret"`
}

type deleteSecretRequest struct {
	Name string `json:"Name"`
}

type wrappedDeleteSecret struct {
	Secret deleteSecretRequest `json:"Secret"`
}

type unveilRequest struct {
	Name    string `json:"Name"`
	Version *int   `json:"Version"`
}

type wrappedUnveilRequest struct {
	Secret unveilRequest `json:"Secret"`
}

type unveilResponse struct {
	Name    string `json:"Name"`
	Version int    `json:"Version"`
	Value   string `json:"Value"`
}

type wrappedUnveilResponse struct {
	Secret unveilResponse `json:"Secret"`
}

type paginatedSecretList struct {
	Count   int              `json:"Count"`
	From    int              `json:"From"`
	Total   int              `json:"Total"`
	Secrets []secretResponse `json:"Secrets"`
}

// Server is the local SecretManager API server.
type Server struct {
	store  *Store
	mux    *http.ServeMux
	prefix string
}

// NewServer creates a new Server with the given path prefix.
// prefix should be like "/api/cloud/1.1" (no trailing slash).
func NewServer(prefix string) *Server {
	s := &Server{
		store:  NewStore(),
		mux:    http.NewServeMux(),
		prefix: prefix,
	}
	s.mux.HandleFunc(prefix+"/", s.handleRequest)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, s.prefix)

	// POST /secretmanager/vaults/{id}/secrets/unveil
	if strings.HasSuffix(path, "/secrets/unveil") && r.Method == http.MethodPost {
		vaultID := extractVaultID(path, "/secrets/unveil")
		if vaultID == "" {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		s.handleUnveil(w, r, vaultID)
		return
	}

	// /secretmanager/vaults/{id}/secrets
	if strings.HasSuffix(path, "/secrets") {
		vaultID := extractVaultID(path, "/secrets")
		if vaultID == "" {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			s.handleListSecrets(w, r, vaultID)
		case http.MethodPost:
			s.handleCreateSecret(w, r, vaultID)
		case http.MethodDelete:
			s.handleDeleteSecret(w, r, vaultID)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	http.NotFound(w, r)
}

// extractVaultID extracts the vault ID from a path like
// /secretmanager/vaults/{id}/secrets by trimming the suffix and prefix.
func extractVaultID(path, suffix string) string {
	trimmed := strings.TrimSuffix(path, suffix)
	const prefix = "/secretmanager/vaults/"
	if !strings.HasPrefix(trimmed, prefix) {
		return ""
	}
	id := strings.TrimPrefix(trimmed, prefix)
	if id == "" || strings.Contains(id, "/") {
		return ""
	}
	return id
}

func (s *Server) handleListSecrets(w http.ResponseWriter, _ *http.Request, vaultID string) {
	secrets := s.store.List(vaultID)
	items := make([]secretResponse, len(secrets))
	for i, sec := range secrets {
		items[i] = secretResponse{Name: sec.Name, LatestVersion: sec.LatestVersion}
	}
	resp := paginatedSecretList{
		Count:   len(items),
		From:    0,
		Total:   len(items),
		Secrets: items,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCreateSecret(w http.ResponseWriter, r *http.Request, vaultID string) {
	var req wrappedCreateSecret
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Secret.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	latestVersion, err := s.store.Create(vaultID, req.Secret.Name, req.Secret.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := wrappedSecret{
		Secret: secretResponse{Name: req.Secret.Name, LatestVersion: latestVersion},
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleDeleteSecret(w http.ResponseWriter, r *http.Request, vaultID string) {
	var req wrappedDeleteSecret
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.store.Delete(vaultID, req.Secret.Name); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUnveil(w http.ResponseWriter, r *http.Request, vaultID string) {
	var req wrappedUnveilRequest
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	version := 0
	if req.Secret.Version != nil {
		version = *req.Secret.Version
	}
	value, actualVersion, err := s.store.Unveil(vaultID, req.Secret.Name, version)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	resp := wrappedUnveilResponse{
		Secret: unveilResponse{
			Name:    req.Secret.Name,
			Version: actualVersion,
			Value:   value,
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

func readJSON(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
