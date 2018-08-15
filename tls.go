package main

import (
	"crypto/tls"
	"path/filepath"
)

func getPeerCertificatePath(dataDir string, fingerprint Fingerprint) string {
	return filepath.Join(dataDir, DataDirTrusted, fingerprint.String())
}

func getPeerTLS(peer Peer) (*tls.Config, error) {
	return nil, nil
}
