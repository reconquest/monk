package main

import (
	"encoding/json"
)

type Serializable interface {
	Serialize() []byte
}

type PacketLight struct {
	Network []string `json:"network"`
}

func (packet PacketLight) Serialize() []byte {
	marshaled, err := json.Marshal(packet)
	if err != nil {
		panic(err)
	}

	return marshaled
}
