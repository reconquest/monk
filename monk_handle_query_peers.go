package main

import (
	"os"
)

func (monk *Monk) HandleQueryPeers(query Packet) (Packetable, error) {
	peers := monk.peers.get()

	for i, peer := range peers {
		if monk.isTrustedPeer(peer.Fingerprint) {
			peers[i].Trusted = true
		}
	}

	return PacketPeers(peers), nil
}

func (monk *Monk) isTrustedPeer(fingerprint Fingerprint) bool {
	_, err := os.Stat(
		getPeerCertificatePath(monk.dataDir, fingerprint),
	)
	return !os.IsNotExist(err)
}
