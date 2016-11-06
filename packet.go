package main

type Packetable interface {
	Signature() []byte
}

type PacketHere struct {
	Network string `json:"net"`
}

func (here PacketHere) Signature() []byte {
	return []byte{'H', 'E', 'R', 'E'}
}
