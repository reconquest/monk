package main

import (
	"crypto/tls"
	"fmt"

	"github.com/reconquest/karma-go"
)

var (
	x = 0
)

func (monk *Monk) HandleQueryStream(query Packet) (Packetable, error) {
	var request PacketQueryStream
	err := query.Bind(&request)
	if err != nil {
		return nil, err
	}

	if request.ID == "" {
		return monk.handleNewStream()
	} else {
		return monk.handleConnectStream(request)
	}
}

func (monk *Monk) handleConnectStream(query PacketQueryStream) (Packetable, error) {
	peer, err := monk.getPeer(query.ID)
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf(
		"%s:%d",
		peer.IP,
		monk.port,
	)

	client := NewClient("tcp", address)

	err = client.Dial()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable dial connection to the peer: %s", address,
		)
	}

	var response PacketEncryptConnection
	err = client.Query(PacketEncryptConnection{
		ID:          monk.machine,
		Fingerprint: monk.security.Fingerprint,
	}, &response)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to query for encryption",
		)
	}

	secured := tls.Client(client.conn, &tls.Config{
		Certificates:       []tls.Certificate{monk.security.X509KeyPair},
		InsecureSkipVerify: true,
	})

	infof("stream: establishing tls encryption with %s", peer.Machine)

	err = secured.Handshake()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to complete handshake with remote server",
		)
	}

	infof("stream: handshake completed with %s", peer.Machine)

	pipe, err := StartPipe(monk.dataDir)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to init pipe",
		)
	}

	go func() {
		err := pipe.WaitConnect()
		if err != nil {
			secured.Close()
			pipe.Close()
			errorh(err, "unable to accept pipe connection")
			return
		}

		communicate(
			pipe, secured,
			secured, pipe,
		)
	}()

	return PacketStream{
		Pipe: pipe.path,
	}, nil
}

func (monk *Monk) handleNewStream() (Packetable, error) {
	pipe, err := StartPipe(monk.dataDir)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to init pipe",
		)
	}

	x++
	go func() {
		err := pipe.WaitConnect()
		if err != nil {
			errorh(err, "unable to wait for pipe connection")
			return
		}

		monk.stream.Serve(
			monk.machine+"-"+fmt.Sprint(x),
			pipe,
		)
	}()

	return PacketStream{
		Pipe: pipe.path,
	}, nil
}
