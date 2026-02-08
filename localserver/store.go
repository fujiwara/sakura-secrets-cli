package localserver

import (
	"crypto/rand"
	"fmt"
	"sort"
	"sync"
)

// SecretMeta represents secret metadata returned by List.
type SecretMeta struct {
	Name          string
	LatestVersion int
}

// Store is an in-memory store for secrets with versioning and simple XOR encryption.
type Store struct {
	mu     sync.RWMutex
	vaults map[string]map[string]*secret // vaultID -> secretName -> secret
}

type secret struct {
	name     string
	versions map[int][]byte // version -> XOR-encrypted value
	latest   int
	key      []byte // XOR key
}

// NewStore creates a new empty Store.
func NewStore() *Store {
	return &Store{
		vaults: make(map[string]map[string]*secret),
	}
}

func (s *Store) getOrCreateVault(vaultID string) map[string]*secret {
	v, ok := s.vaults[vaultID]
	if !ok {
		v = make(map[string]*secret)
		s.vaults[vaultID] = v
	}
	return v
}

// List returns all secret metadata for a vault.
func (s *Store) List(vaultID string) []SecretMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vault, ok := s.vaults[vaultID]
	if !ok {
		return nil
	}
	result := make([]SecretMeta, 0, len(vault))
	for _, sec := range vault {
		result = append(result, SecretMeta{
			Name:          sec.name,
			LatestVersion: sec.latest,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Create creates a new secret or adds a new version to an existing secret.
// Returns the latest version number.
func (s *Store) Create(vaultID, name, value string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vault := s.getOrCreateVault(vaultID)
	sec, ok := vault[name]
	if !ok {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return 0, fmt.Errorf("failed to generate encryption key: %w", err)
		}
		sec = &secret{
			name:     name,
			versions: make(map[int][]byte),
			key:      key,
		}
		vault[name] = sec
	}
	sec.latest++
	sec.versions[sec.latest] = xorEncrypt([]byte(value), sec.key)
	return sec.latest, nil
}

// Unveil decrypts and returns a secret value.
// If version is 0, the latest version is returned.
func (s *Store) Unveil(vaultID, name string, version int) (string, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vault, ok := s.vaults[vaultID]
	if !ok {
		return "", 0, fmt.Errorf("secret %q not found", name)
	}
	sec, ok := vault[name]
	if !ok {
		return "", 0, fmt.Errorf("secret %q not found", name)
	}
	if version == 0 {
		version = sec.latest
	}
	encrypted, ok := sec.versions[version]
	if !ok {
		return "", 0, fmt.Errorf("secret %q version %d not found", name, version)
	}
	return string(xorEncrypt(encrypted, sec.key)), version, nil
}

// Delete removes a secret from a vault.
func (s *Store) Delete(vaultID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vault, ok := s.vaults[vaultID]
	if !ok {
		return fmt.Errorf("secret %q not found", name)
	}
	if _, ok := vault[name]; !ok {
		return fmt.Errorf("secret %q not found", name)
	}
	delete(vault, name)
	return nil
}

// xorEncrypt XORs data with key (repeating key as needed).
func xorEncrypt(data, key []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[i] = b ^ key[i%len(key)]
	}
	return result
}
