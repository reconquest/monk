package main

import (
	"net"
	"strconv"
	"sync"

	"github.com/reconquest/ser-go"
)

var (
	packetMaxSize = 1024
	packetPrefix  = []byte{'C', 'U', 'R', 'E'}
)

type Peer struct {
	Source *net.IPNet
	Peers  []*Peer
}

type Monk struct {
	*sync.Mutex

	port int
	udp  net.PacketConn

	networks []*net.IPNet

	peers   []Peer
	seeds   []Peer
	seeding bool
}

func NewMonk(port int) *Monk {
	return &Monk{
		Mutex: &sync.Mutex{},
		port:  port,
	}
}

func (monk *Monk) addNetwork(network *net.IPNet) {
	monk.networks = append(monk.networks, network)
}

func (monk *Monk) bind() error {
	udp, err := net.ListenPacket("udp", ":"+strconv.Itoa(monk.port))
	if err != nil {
		return err
	}

	infof("listening at :%d", monk.port)

	monk.udp = udp

	return nil
}

func (monk *Monk) observe() {
	assert(
		monk.udp != nil,
		"observing network without connection",
	)

	for {
		remote, packet, err := monk.read()
		if err != nil {
			errorh(err, "unable to read packet")

			continue
		}

		go monk.handle(remote, packet)
	}
}

func (monk *Monk) heartbeat() {
	monk.broadcastLight()
}

func (monk *Monk) broadcastLight() {
	network := []string{}
	for _, ipnet := range monk.networks {
		network = append(network, ipnet.Network())
	}

	packet := PacketLight{
		Network: network,
	}

	err := monk.broadcast(packet)
	if err != nil {
		errorh(err, "can't broadcast packet: %v", packet)
		return
	}
}

func (monk *Monk) read() (net.Addr, []byte, error) {
	packet := make([]byte, packetMaxSize)

	size, remote, err := monk.udp.ReadFrom(packet)
	if err != nil {
		return nil, nil, err
	}

	return remote, packet[:size], nil
}

func (monk *Monk) handle(remote net.Addr, packet []byte) error {
	debugf("remote: %s; packet: %s", remote, string(packet))

	return nil
}

func (monk *Monk) broadcast(packet Serializable) error {
	data := pack(packet)

	for _, network := range monk.networks {
		address := getBroadcastIP(network)

		debugf("broadcasting to %s", address)

		_, err := monk.udp.WriteTo(
			data,
			&net.UDPAddr{IP: address, Port: monk.port},
		)
		if err != nil {
			return ser.Errorf(
				err, "can't broadcast to %s:%d (network: %s)",
				address, monk.port, network,
			)
		}
	}

	return nil
}

func pack(data Serializable) []byte {
	serialized := data.Serialize()

	size := []byte(strconv.Itoa(len(serialized)))

	packet := []byte{}
	packet = append(packet, packetPrefix...)
	packet = append(packet, byte(0))
	packet = append(packet, size...)
	packet = append(packet, byte(0))
	packet = append(packet, serialized...)

	return packet
}
