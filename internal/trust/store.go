package trust

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
)

var ErrNotFound = errors.New("trust CA not found")

const (
	caFilename   = "ca.pem"
	modeFilename = "mode"
)

// Mode records how TLS trust was established for a CLI namespace.
type Mode string

const (
	ModeNamespace Mode = "namespace"
	ModeSystem    Mode = "system"
	ModeInsecure  Mode = "insecure"
)

func trustDir(namespace string) (string, error) {
	root := config.ConfigFolder()
	if root == "" {
		return "", fmt.Errorf("config folder is not initialized")
	}
	return filepath.Join(root, "trust", namespace), nil
}

func caPath(namespace string) (string, error) {
	dir, err := trustDir(namespace)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, caFilename), nil
}

func modePath(namespace string) (string, error) {
	dir, err := trustDir(namespace)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, modeFilename), nil
}

// StoreCA persists spec.ca (base64 PEM) for a CLI namespace.
func StoreCA(namespace, caBase64 string) error {
	if strings.TrimSpace(caBase64) == "" {
		return nil
	}
	pem, err := base64.StdEncoding.DecodeString(caBase64)
	if err != nil {
		return fmt.Errorf("decode trust CA: %w", err)
	}
	dir, err := trustDir(namespace)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path, err := caPath(namespace)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, pem, 0600); err != nil {
		return err
	}
	return SetCachedMode(namespace, ModeNamespace)
}

// GetCA returns the stored PEM for a CLI namespace.
func GetCA(namespace string) ([]byte, error) {
	path, err := caPath(namespace)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

// HasCA reports whether a namespace trust CA file exists.
func HasCA(namespace string) bool {
	path, err := caPath(namespace)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// RemoveCA deletes stored trust material for a namespace (3G delete scope).
func RemoveCA(namespace string) error {
	dir, err := trustDir(namespace)
	if err != nil {
		return err
	}
	err = os.RemoveAll(dir)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// GetCachedMode returns a previously cached TLS trust mode, if any.
func GetCachedMode(namespace string) (Mode, bool) {
	path, err := modePath(namespace)
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	mode := Mode(strings.TrimSpace(string(data)))
	switch mode {
	case ModeSystem, ModeInsecure, ModeNamespace:
		return mode, true
	default:
		return "", false
	}
}

// SetCachedMode persists probe outcome for a namespace.
func SetCachedMode(namespace string, mode Mode) error {
	dir, err := trustDir(namespace)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path, err := modePath(namespace)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(string(mode)), 0600)
}
