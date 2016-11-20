package main

import "encoding/json"

type Serializable interface {
	Serialize() []byte
}

type PacketPresence struct {
	Networks []string `json:"networks"`
}

func (packet PacketPresence) Serialize() []byte {
	marshaled, err := json.Marshal(packet)
	if err != nil {
		panic(err)
	}

	return marshaled
}
