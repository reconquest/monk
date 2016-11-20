package main

import (
	"encoding/binary"
	"net"
)

type Networks []Network

func (networks Networks) Strings() []string {
	strings := []string{}
	for _, network := range networks {
		strings = append(strings, network.String())
	}
	return strings
}

type Network struct {
	*net.IPNet
}

func (network Network) String() string {
	return network.IPNet.String()
}

func (network *Network) getBroadcastAddress() net.IP {
	ip := make(net.IP, len(network.IP.To4()))
	binary.BigEndian.PutUint32(
		ip,
		binary.BigEndian.Uint32(network.IP.To4())|^binary.BigEndian.Uint32(
			net.IP(network.Mask).To4(),
		),
	)
	return ip
}
