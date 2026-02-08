package localserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	base := prefix + "/secretmanager/vaults/{vault_id}"
	s.mux.HandleFunc("GET "+base+"/secrets", s.handleListSecrets)
	s.mux.HandleFunc("POST "+base+"/secrets", s.handleCreateSecret)
	s.mux.HandleFunc("DELETE "+base+"/secrets", s.handleDeleteSecret)
	s.mux.HandleFunc("POST "+base+"/secrets/unveil", s.handleUnveil)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleListSecrets(w http.ResponseWriter, r *http.Request) {
	vaultID := r.PathValue("vault_id")
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

func (s *Server) handleCreateSecret(w http.ResponseWriter, r *http.Request) {
	vaultID := r.PathValue("vault_id")
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

func (s *Server) handleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	vaultID := r.PathValue("vault_id")
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

func (s *Server) handleUnveil(w http.ResponseWriter, r *http.Request) {
	vaultID := r.PathValue("vault_id")
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
