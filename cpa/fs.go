package cpa

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// LoadPolicyFromFS takes a filesystem path to load policy files from. It returns a parsed policy.
// If the path is a file that policy is loaded as a bundle of 1 file. If the path is a directory that
// directory is walked recursively searching for all rego files. If the bundle is empty an error is returned.
func LoadPolicyFromFS(root string) (*Policy, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".rego" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk root: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no rego policies found at path: %q", root)
	}

	bundle := make(map[string]string)
	for _, file := range files {
		data, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		bundle[file] = string(data)
	}

	return ParseBundle(bundle)
}
