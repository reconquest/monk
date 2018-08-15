package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/reconquest/karma-go"
)

func (monk *Monk) HandleTCPPacket(conn *TCPConnection, query Packet) (Packetable, error) {
	logger.Debugf("socket: %s", query.Signature)

	switch query.Signature {
	case SignatureQueryCertificate:
		return monk.HandleQueryCertificate(query)

	case SignatureEncryptConnection:
		return monk.HandleEncryptConnection(conn, query)

	default:
		return nil, fmt.Errorf(
			"unexpected packet signature: %s",
			query.Signature,
		)
	}
}

func (monk *Monk) HandleTCPConnection(conn *TCPConnection) error {
	if conn.peer == nil {
		return fmt.Errorf("conn.peer is nil")
	}

	certPath := getPeerCertificatePath(monk.dataDir, conn.peer.Fingerprint)

	certData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return karma.Format(
			err,
			"unable to read certificate data of peer: %s", conn.peer.Machine,
		)
	}

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(certData)
	if !ok {
		return fmt.Errorf(
			"unable to add client certificate to tls.Server pool: %s",
			conn.peer.Machine,
		)
	}

	secured := tls.Server(
		conn.conn,
		&tls.Config{
			Certificates: []tls.Certificate{monk.security.X509KeyPair},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    pool,
		},
	)

	worker := monk.stream.Serve(conn.peer.Machine, secured)
	worker.Wait()

	return nil
}
