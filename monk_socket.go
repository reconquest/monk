package main

import (
	"fmt"
)

func (monk *Monk) HandleSock(query Packet) (Packetable, error) {
	logger.Debugf("socket: %s", query.Signature)

	switch query.Signature {
	case SignatureQueryPeers:
		return monk.HandleQueryPeers(query)

	case SignatureTrustPeer:
		return monk.HandleTrustPeer(query)

	case SignatureQueryStream:
		return monk.HandleQueryStream(query)

	default:
		return nil, fmt.Errorf(
			"unexpected packet signature: %s",
			query.Signature,
		)
	}
}
