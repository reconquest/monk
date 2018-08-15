package main

import (
	"encoding/json"

	"github.com/reconquest/karma-go"
)

type PacketHandler func(Packet) (Packetable, error)

type Packetable interface {
	Signature() string
}

type Packet struct {
	Signature string          `json:"signature"`
	Payload   json.RawMessage `json:"payload"`
}

func makePacket(data Packetable) interface{} {
	return struct {
		Signature string      `json:"signature"`
		Payload   interface{} `json:"payload"`
	}{data.Signature(), data}
}

func pack(data Packetable) []byte {
	encoded, err := json.Marshal(makePacket(data))
	if err != nil {
		panic(
			karma.Format(
				err,
				"unable to encode packet: %s", data.Signature(),
			),
		)
	}

	return encoded
}

func unpack(data []byte) (Packet, error) {
	var packet Packet
	err := json.Unmarshal(data, &packet)
	if err != nil {
		return packet, karma.Format(
			err,
			"unable to decode packet",
		)
	}

	return packet, nil
}

func (packet *Packet) Bind(data Packetable) error {
	err := json.Unmarshal(packet.Payload, data)
	if err != nil {
		return karma.Format(
			err,
			"unable to decode packet payload",
		)
	}

	return nil
}
