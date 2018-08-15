package main

func (monk *Monk) HandleEncryptConnection(conn *TCPConnection, query Packet) (Packetable, error) {
	var request PacketEncryptConnection
	err := query.Bind(&request)
	if err != nil {
		return nil, err
	}

	peer, err := monk.getPeer(request.ID)
	if err != nil {
		return PacketError{
			"unknown peer",
		}, nil
	}

	if !monk.isTrustedPeer(peer.Fingerprint) {
		return PacketError{
			"untrusted peer",
		}, nil
	}

	conn.peer = peer

	return PacketEncryptConnection{}, nil
}
