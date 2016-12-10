package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"

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

	networks Networks

	peers Peers
}

func NewMonk(port int) *Monk {
	monk := &Monk{
		mutex: &sync.Mutex{},
		port:  port,
		peers: Peers{
			mutex: &sync.Mutex{},
		},
	}

	return monk
}

func (monk *Monk) addNetwork(network Network) {
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
		monk.peers.cleanup(heartbeatInterval * 2)
	})

	// try to find network that does not have a presencer

	go monk.broadcastPresence(monk.networks)
}

func (monk *Monk) broadcastPresence(networks Networks) {
	packet := PacketPresence{
		"key": "value",
	}

	for _, network := range networks {
		go monk.broadcast(network, packet)

	}
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

	var peer *Peer
	var ok bool
	if peer, ok = monk.peers.find(remote.IP, remote.Network()); !ok {
		peer = &Peer{
			ip:      remote.IP,
			network: remote.Network(),
		}

		monk.peers.add(peer)
	}

	peer.data = presence
	peer.last = time.Now()

	return
}

func (monk *Monk) broadcast(network Network, packet Serializable) {
	data := pack(packet)

	address := network.getBroadcastAddress()

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
