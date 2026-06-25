package config

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// migrateConfigPermissions tightens permissions on an existing CLI config tree.
func migrateConfigPermissions(root string) error {
	if root == "" {
		return nil
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer r.Close()

	return migrateDirPermissions(r, ".")
}

func migrateDirPermissions(r *os.Root, rel string) error {
	if err := chmodRootEntry(r, rel); err != nil {
		return err
	}
	entries, err := fs.ReadDir(r.FS(), rel)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		childRel := joinRel(rel, entry.Name())
		if err := chmodRootEntry(r, childRel); err != nil {
			return err
		}
		if entry.IsDir() {
			if err := migrateDirPermissions(r, childRel); err != nil {
				return err
			}
		}
	}
	return nil
}

func joinRel(parent, name string) string {
	if parent == "." {
		return name
	}
	return filepath.Join(parent, name)
}

func chmodRootEntry(r *os.Root, rel string) error {
	f, err := r.Open(filepath.ToSlash(rel))
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	mode := info.Mode()
	if mode&os.ModeSymlink != 0 {
		return nil
	}
	target := os.FileMode(util.FilePerm)
	if info.IsDir() {
		target = os.FileMode(util.DirPerm)
	} else if mode.Perm()&0111 != 0 {
		target = os.FileMode(util.ExecPerm)
	}
	if mode.Perm()&0777 == target&0777 {
		return nil
	}
	return f.Chmod(target)
}
