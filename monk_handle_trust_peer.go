package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/reconquest/karma-go"
)

func (monk *Monk) HandleTrustPeer(query Packet) (Packetable, error) {
	var toTrust PacketTrustPeer
	err := query.Bind(&toTrust)
	if err != nil {
		return nil, err
	}

	peer, err := monk.getPeer(toTrust.ID)
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf("%s:%d", peer.IP, monk.port)

	client := NewClient("tcp", address)

	err = client.Dial()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable dial connection to the peer: %s", address,
		)
	}

	defer client.Close()

	var certPEM PacketCertificate
	err = client.Query(PacketQueryCertificate{}, &certPEM)
	if err != nil {
		return nil, err
	}

	certBlock, _ := pem.Decode(certPEM.Data)
	if certBlock == nil {
		return nil, karma.Format(
			err,
			"unable to decode peer's certificate PEM data",
		)
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to parse certificate PEM block",
		)
	}

	fingerprint := getFingerprint(cert)

	if bytes.Compare(fingerprint, peer.Fingerprint) != 0 {
		return nil, karma.
			Describe("known", peer.Fingerprint.String()).
			Describe("got", fingerprint.String()).
			Format(
				err,
				"the peer provided certificate with different fingerprint",
			)
	}

	err = ioutil.WriteFile(
		getPeerCertificatePath(monk.dataDir, fingerprint),
		certPEM.Data,
		0600,
	)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to save certificate PEM block in data dir",
		)
	}

	return PacketPeer{peer}, nil
}
