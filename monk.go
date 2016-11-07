package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/reconquest/ser-go"
)

var (
	packetBufferReadSize = 1024 * 4
	packetPrefix         = []byte{'M', 'O', 'N', 'K', 0}
)

type Monk struct {
	mutex *sync.Mutex

	port int
	udp  net.PacketConn

	networks []*net.IPNet

	peers      Peers
	presencers Peers
}

func NewMonk(port int) *Monk {
	return &Monk{
		mutex: &sync.Mutex{},
		port:  port,
		peers: Peers{
			mutex: &sync.Mutex{},
		},
		presencers: Peers{
			mutex: &sync.Mutex{},
		},
	}
}

func (monk *Monk) addNetwork(network *net.IPNet) {
	monk.networks = append(monk.networks, network)
}

func (monk *Monk) bind() error {
	udp, err := net.ListenUDP("udp", &net.UDPAddr{Port: monk.port})
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

		go monk.handle(remote.(*net.UDPAddr), packet)
	}
}

func (monk *Monk) heartbeat() {
	monk.withLock(func() {
		monk.presencers.cleanup(heartbeatInterval)
	})

	// try to find network that does not have presencer

	networks := []string{}
	for _, peer := range monk.presencers {
		found := false
		for _, network := range networks {
			if network == peer.network {
				found = true
				break
			}
		}

		if !found {
			networks = append(networks, peer.network)
		}
	}

	networksOrphans := []string{}
	for _, network := range monk.networks {
		presencing := false
		for _, presencingNetwork := range networks {
			if presencingNetwork == network.String() {

			}
		}
	}

	monk.broadcastPresence()
}

func (monk *Monk) broadcastPresence(target) {
	network := []string{}
	for _, ipnet := range monk.networks {
		network = append(network, ipnet.Network())
	}

	packet := PacketPresence{
		Network: network,
	}

	monk.broadcast(packet)
}

func (monk *Monk) read() (net.Addr, []byte, error) {
	packet := make([]byte, packetBufferReadSize)

	size, remote, err := monk.udp.ReadFrom(packet)
	if err != nil {
		return nil, nil, err
	}

	return remote, packet[:size], nil
}

func (monk *Monk) handle(remote *net.UDPAddr, data []byte) {
	mine := false
	for _, network := range monk.networks {
		if bytes.Compare(remote.IP, network.IP) == 0 {
			mine = true
			break
		}
	}

	if mine {
		debugf("mine packet, skipping")
		return
	}

	packet, err := unpack(data)
	if err != nil {
		errorh(
			err,
			"unable to unpack packet: %s", data,
		)
		return
	}

	// TODO: some day here can be not only presence, so be aware here can be
	// panic if unpack returned not PacketPresence
	presence := packet.(PacketPresence)
	fmt.Printf("XXXXXX monk.go:136 presence: %#v\n", presence)

	return
}

func (monk *Monk) broadcast(network *net.IPNet, packet Serializable) {
	data := pack(packet)

	address := getBroadcastIP(network)

	debugf("broadcasting to %s", address)

	_, err := monk.udp.WriteTo(
		data,
		&net.UDPAddr{IP: address, Port: monk.port},
	)
	if err != nil {
		errorh(
			err,
			"unable to send packet to %s:%d (network: %s)",
			address, monk.port, network.Network(),
		)
		return
	}
}

func pack(data Serializable) []byte {
	return append(packetPrefix, data.Serialize()...)
}

func unpack(packet []byte) (Serializable, error) {
	if !bytes.HasPrefix(packet, append(packetPrefix, byte(0))) {
		return nil, errors.New("packet does not contain specific signature")
	}

	packet = packet[len(packetPrefix):]

	var presence PacketPresence

	err := json.Unmarshal(packet, &presence)
	if err != nil {
		return nil, ser.Errorf(
			err, "unable to unmarshal packet data",
		)
	}

	return presence, nil
}

func (monk *Monk) withLock(fun func()) {
	monk.mutex.Lock()
	defer monk.mutex.Unlock()

	fun()
}
