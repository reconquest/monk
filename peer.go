package main

import (
	"encoding/json"
	"net"
	"strconv"

	"github.com/reconquest/ser-go"
)

var (
	packetMaxSize = 1024
	packetPrefix  = []byte{'C', 'U', 'R', 'E'}
)

type Peer struct {
	networks   []*net.IPNet
	port       int
	connection net.PacketConn
}

func NewPeer(port int) *Peer {
	return &Peer{port: port}
}

func (peer *Peer) addNetwork(network *net.IPNet) {
	peer.networks = append(peer.networks, network)
}

func (peer *Peer) bind() error {
	connection, err := net.ListenPacket("udp", ":"+strconv.Itoa(peer.port))
	if err != nil {
		return err
	}

	infof("listening at :%d", peer.port)

	peer.connection = connection

	return nil
}

func (peer *Peer) observe() {
	assert(
		peer.connection != nil,
		"observing network without connection",
	)

	for {
		remote, packet, err := peer.read()
		if err != nil {
			errorh(err, "unable to read packet")

			continue
		}

		go peer.handle(remote, packet)
	}
}

func (peer *Peer) heartbeat() {
	assert(
		peer.connection != nil,
		"heartbeat without network connection",
	)

	packet := PacketHere{}

	data, err := serialize(packet)
	if err != nil {
		errorh(err, "can't serialize packet: %#v", packet)

		return
	}

	err = peer.broadcast(data)
	if err != nil {
		errorh(err, "can't broadcast packet: %v", data)
		return
	}
}

func (peer *Peer) read() (net.Addr, []byte, error) {
	packet := make([]byte, packetMaxSize)

	size, remote, err := peer.connection.ReadFrom(packet)
	if err != nil {
		return nil, nil, err
	}

	return remote, packet[:size], nil
}

func (peer *Peer) handle(remote net.Addr, packet []byte) error {
	debugf("remote: %s; packet: %s", remote, string(packet))

	return nil
}

func (peer *Peer) broadcast(packet []byte) error {
	for _, network := range peer.networks {
		address := getBroadcastIP(network)

		debugf("broadcasting to %s", address)

		_, err := peer.connection.WriteTo(
			packet,
			&net.UDPAddr{IP: address, Port: peer.port},
		)
		if err != nil {
			return ser.Errorf(
				err, "can't broadcast to %s:%d (network: %s)",
				address, peer.port, network,
			)
		}
	}

	return nil
}

func serialize(packet Packetable) ([]byte, error) {
	marshaled, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}

	body := []byte{}
	body = append(body, packet.Signature()...)
	body = append(body, byte(0))
	body = append(body, marshaled...)

	size := []byte(strconv.Itoa(len(body)))

	data := []byte{}
	data = append(data, packetPrefix...)
	data = append(data, byte(0))
	data = append(data, size...)
	data = append(data, byte(0))
	data = append(data, body...)

	return data, nil
}
