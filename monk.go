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

	networks Networks

	peers       Peers
	presencers  Peers
	connections Peers

	// TODO: try to avoid this
	limits struct {
		connections struct {
			min int
			max int
		}
	}
}

func NewMonk(port, min, max int) *Monk {
	monk := &Monk{
		mutex: &sync.Mutex{},
		port:  port,
		peers: Peers{
			mutex: &sync.Mutex{},
		},
		presencers: Peers{
			mutex: &sync.Mutex{},
		},
	}

	monk.limits.connections.min = min
	monk.limits.connections.max = max

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
	var (
		beatingNetworks = []string{}
		silentNetworks  = []Network{}
	)

	monk.withLock(func() {
		monk.presencers.cleanup(heartbeatInterval * 2)

		beatingNetworks = monk.presencers.getNetworks()
	})

	// try to find network that does not have a presencer

	for _, network := range monk.networks {
		beating := false
		for _, beatingNetwork := range beatingNetworks {
			if network.String() == beatingNetwork {
				beating = true
				break
			}
		}

		if !beating {
			silentNetworks = append(silentNetworks, network)
		}
	}

	go monk.broadcastPresence(silentNetworks)
}

func (monk *Monk) broadcastPresence(networks Networks) {
	packet := PacketPresence{
		Networks: monk.networks.Strings(),
		Peers:    monk.getPeersTable(),
	}

	for _, network := range networks {
		go monk.broadcast(network, packet)

	}
}

func (monk *Monk) getPeersTable() map[string][]string {
	table := map[string][]string{}
	for _, peer := range monk.peers.peers {
		for _, network := range peer.networks {
			_, ok := table[network]
			if !ok {
				table[network] = []string{peer.ip}
			} else {
				table[network] = append(table[network], peer.ip)
			}
		}
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
	fmt.Printf("XXXXXX monk.go:136 presence: %#v\n", presence)

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
