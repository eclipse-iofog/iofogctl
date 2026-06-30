package wasm

import (
	"io"
	"os"
	"path/filepath"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// InstallLocalBinaries copies staged WASM shims into destDir with mode 0755.
func InstallLocalBinaries(staged []StagedBinary, destDir string) ([]string, error) {
	if destDir == "" {
		destDir = defaultInstallDir
	}
	if len(staged) == 0 {
		return nil, nil
	}

	installed := make([]string, 0, len(staged))
	for _, item := range staged {
		if item.LocalPath == "" || item.CanonicalName == "" {
			continue
		}
		destPath := filepath.Join(destDir, item.CanonicalName)
		if err := installBinary(item.LocalPath, destPath); err != nil {
			return installed, err
		}
		installed = append(installed, destPath)
	}
	return installed, nil
}

func installBinary(srcPath, destPath string) error {
	src, err := util.OpenValidatedFile(srcPath)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(src)

	if err := os.MkdirAll(filepath.Dir(destPath), util.DirPerm); err != nil {
		return err
	}

	dst, err := util.CreateUserFile(destPath, util.ExecPerm)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(dst)

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}
