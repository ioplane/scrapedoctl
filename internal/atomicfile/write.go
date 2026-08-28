// Package atomicfile provides durable same-directory file replacement.
package atomicfile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ErrSymlink is returned when the destination is a symbolic link.
var ErrSymlink = errors.New("atomic destination must not be a symbolic link")

// Replace writes data to a same-directory temporary file, syncs it, and atomically renames it.
func Replace(path string, mode fs.FileMode, data []byte) error {
	if err := rejectSymlink(path); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	temporary, createErr := os.CreateTemp(dir, "."+filepath.Base(path)+"-*")
	if createErr != nil {
		return fmt.Errorf("create temporary file: %w", createErr)
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()        //nolint:gosec // Best-effort cleanup after the primary operation.
		_ = os.Remove(temporaryPath) //nolint:gosec // Best-effort cleanup after the primary operation.
	}()

	if err := temporary.Chmod(mode.Perm()); err != nil {
		return fmt.Errorf("set temporary file mode: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		return fmt.Errorf("write temporary file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync temporary file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	if err := rejectSymlink(path); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace destination: %w", err)
	}
	if err := os.Chmod(path, mode.Perm()); err != nil {
		return fmt.Errorf("enforce destination mode: %w", err)
	}

	directory, openErr := os.Open(dir)
	if openErr != nil {
		return fmt.Errorf("open destination directory: %w", openErr)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync destination directory: %w", err)
	}

	return nil
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("inspect destination: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: %s", ErrSymlink, path)
	}
	return nil
}
