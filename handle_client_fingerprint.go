package main

import (
	"fmt"

	"github.com/reconquest/karma-go"
)

func handleClientFingerprint(dataDir string) error {
	err := ensureDataDir(dataDir)
	if err != nil {
		return karma.Format(
			err,
			"unable to ensure data dir exists",
		)
	}

	security, err := getSecureLayer(dataDir, false)
	if err != nil {
		return err
	}

	fmt.Println(security.Fingerprint.String())

	return nil
}
