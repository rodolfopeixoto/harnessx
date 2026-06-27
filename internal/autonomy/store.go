package autonomy

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const ActiveFileRel = ".harness/config/autonomy"

func Save(root string, level Level) error {
	if root == "" {
		return errors.New("autonomy: empty root")
	}
	if _, err := Gate(level, OpRead); err != nil {
		return err
	}
	path := filepath.Join(root, ActiveFileRel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(string(level)+"\n"), 0o644)
}

func Load(root string) (Level, error) {
	if root == "" {
		return Manual, errors.New("autonomy: empty root")
	}
	data, err := os.ReadFile(filepath.Join(root, ActiveFileRel))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Manual, nil
		}
		return Manual, err
	}
	level := Level(strings.TrimSpace(string(data)))
	if _, err := Gate(level, OpRead); err != nil {
		return Manual, err
	}
	return level, nil
}
