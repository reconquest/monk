package main

import (
	"os"
	"path/filepath"
)

const (
	DataDirStream  = "stream"
	DataDirTrusted = "trusted"
	DataDirTLS     = "tls"
)

func ensureDataDir(base string) error {
	dirs := []string{
		DataDirStream,
		DataDirTrusted,
		DataDirTLS,
	}

	for _, dir := range dirs {
		fullDir := filepath.Join(base, dir)

		if _, err := os.Stat(fullDir); os.IsNotExist(err) {
			err = os.MkdirAll(fullDir, 0700)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
