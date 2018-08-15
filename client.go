package main

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/reconquest/karma-go"
)

type Client struct {
	network string
	address string
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder

	sync.Mutex
}

func NewClient(network, address string) *Client {
	client := &Client{}
	client.network = network
	client.address = address
	return client
}

func (client *Client) Close() error {
	return client.conn.Close()
}

func (client *Client) Dial() error {
	conn, err := net.Dial(client.network, client.address)
	if err != nil {
		return err
	}

	client.Lock()
	client.conn = conn
	client.decoder = json.NewDecoder(conn)
	client.encoder = json.NewEncoder(conn)
	client.Unlock()

	return nil
}

func (client *Client) Query(query Packetable, reply Packetable) error {
	err := client.encoder.Encode(makePacket(query))
	if err != nil {
		return karma.Format(
			err,
			"unable to write packet",
		)
	}

	var raw Packet
	err = client.decoder.Decode(&raw)
	if err != nil {
		return karma.Format(
			err,
			"unable to read daemon reply",
		)
	}

	if raw.Signature == SignatureError {
		var replyErr PacketError
		err := raw.Bind(&replyErr)
		if err != nil {
			return karma.Format(
				err,
				"packet signature is error, but can't bind to error struct",
			)
		}

		return karma.Format(
			replyErr.Error,
			"the daemon returned an error",
		)
	}

	err = raw.Bind(reply)
	if err != nil {
		return karma.Format(
			err,
			"unable to bind reply as %s", reply.Signature(),
		)
	}

	return nil
}
