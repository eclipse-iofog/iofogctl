package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoSystemctlInInstallGoSources(t *testing.T) {
	err := filepath.Walk(".", func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), "systemctl") {
			t.Errorf("%s must not contain systemctl (P9-HYG-1)", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk install package: %v", err)
	}
}
