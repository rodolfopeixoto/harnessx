// SPDX-License-Identifier: MIT

package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type EncryptedFileBackend struct {
	Root string
}

func (e *EncryptedFileBackend) Name() string    { return "encrypted_file" }
func (e *EncryptedFileBackend) Available() bool { return true }

func (e *EncryptedFileBackend) Get(name string) (string, error) {
	store, err := e.read()
	if err != nil {
		return "", err
	}
	if v, ok := store[name]; ok && v != "" {
		return v, nil
	}
	return "", ErrNotFound
}

func (e *EncryptedFileBackend) Set(name, value string) error {
	store, err := e.read()
	if err != nil {
		return err
	}
	if store == nil {
		store = map[string]string{}
	}
	store[name] = value
	return e.write(store)
}

func (e *EncryptedFileBackend) Delete(name string) error {
	store, err := e.read()
	if err != nil {
		return err
	}
	if _, ok := store[name]; !ok {
		return ErrNotFound
	}
	delete(store, name)
	return e.write(store)
}

func (e *EncryptedFileBackend) List() ([]string, error) {
	store, err := e.read()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(store))
	for k := range store {
		names = append(names, k)
	}
	sort.Strings(names)
	return names, nil
}

func (e *EncryptedFileBackend) path() string {
	root := e.Root
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".harness")
	}
	return filepath.Join(root, "secrets.enc")
}

func (e *EncryptedFileBackend) keyPath() string {
	root := e.Root
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".harness")
	}
	return filepath.Join(root, "secret-seed")
}

func (e *EncryptedFileBackend) key() ([]byte, error) {
	path := e.keyPath()
	if data, err := os.ReadFile(path); err == nil && len(data) >= 32 {
		sum := sha256.Sum256(data)
		return sum[:], nil
	}
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, seed, 0o600); err != nil {
		return nil, err
	}
	sum := sha256.Sum256(seed)
	return sum[:], nil
}

func (e *EncryptedFileBackend) read() (map[string]string, error) {
	data, err := os.ReadFile(e.path())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	if len(data) < 12 {
		return nil, errors.New("secrets: encrypted file too short")
	}
	key, err := e.key()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("secrets: decrypt: %w", err)
	}
	var store map[string]string
	if err := json.Unmarshal(plain, &store); err != nil {
		return nil, err
	}
	return store, nil
}

func (e *EncryptedFileBackend) write(store map[string]string) error {
	key, err := e.key()
	if err != nil {
		return err
	}
	plain, err := json.Marshal(store)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext := gcm.Seal(nonce, nonce, plain, nil)
	if err := os.MkdirAll(filepath.Dir(e.path()), 0o700); err != nil {
		return err
	}
	return os.WriteFile(e.path(), ciphertext, 0o600)
}
