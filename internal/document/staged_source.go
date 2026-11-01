package document

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// stageSource performs raw filesystem staging only. Each command retains its
// own candidate validation; rollback must also accept original bytes verbatim.
func stageSource(path string, source []byte, mode os.FileMode) (temporaryPath string, err error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".datapack-stage-*")
	if err != nil {
		return "", err
	}
	temporaryPath = file.Name()
	closed := false
	defer func() {
		if !closed {
			err = errors.Join(err, file.Close())
		}
		if err != nil {
			if cleanup := os.Remove(temporaryPath); cleanup != nil && !os.IsNotExist(cleanup) {
				err = errors.Join(err, cleanup)
			}
		}
	}()
	if _, err = file.Write(source); err != nil {
		return temporaryPath, err
	}
	if err = file.Chmod(mode & (os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky)); err != nil {
		return temporaryPath, err
	}
	if err = file.Sync(); err != nil {
		return temporaryPath, err
	}
	err = file.Close()
	closed = true
	return temporaryPath, err
}

func restoreStagedSource(path string, original []byte, mode os.FileMode, rename func(string, string) error) error {
	temporary, err := stageSource(path, original, mode)
	if err != nil {
		return err
	}
	if err = rename(temporary, path); err != nil {
		// Keep the only on-disk recovery copy when compensation itself fails.
		// The extension is not .cue, so project scanning will not consume it.
		return fmt.Errorf("restore source failed; recovery bytes retained at %q: %w", temporary, err)
	}
	return nil
}
