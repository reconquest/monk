package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/reconquest/karma-go"
)

const (
	intervalPresence = time.Second * 2
)

type Monk struct {
	mutex *sync.Mutex

	machine  string
	security *SecureLayer
	port     int
	udp      net.PacketConn
	dataDir  string

	networks Networks

	peers Peers

	stream           *Stream
	streamBufferSize int
}

func NewMonk(
	machineID string,
	security *SecureLayer,
	dataDir string,
	port int,
	streamBufferSize int,
) *Monk {
	monk := &Monk{
		machine:          machineID,
		security:         security,
		mutex:            &sync.Mutex{},
		dataDir:          dataDir,
		port:             port,
		streamBufferSize: streamBufferSize,
		peers: Peers{
			mutex: &sync.Mutex{},
		},
		stream: NewStream(streamBufferSize),
	}

	return monk
}

func (monk *Monk) Close() {
	monk.stream.Close()
}

func (monk *Monk) SetNetworks(networks []Network) {
	monk.mutex.Lock()
	defer monk.mutex.Unlock()

	same := func(a, b Network) bool {
		if a.Interface == b.Interface &&
			a.BroadcastAddress.String() == b.BroadcastAddress.String() &&
			a.IP.String() == b.IP.String() {
			return true
		}

		return false
	}

	toAdd := []Network{}
	for _, network := range networks {
		found := false
		for _, saved := range monk.networks {
			if same(saved, network) {
				found = true
				break
			}
		}

		if found {
			continue
		}

		toAdd = append(toAdd, network)

		logger.Infof(
			"network added: %s dev %s",
			network.BroadcastAddress,
			network.Interface,
		)
	}

	for index := len(monk.networks) - 1; index > 0; index-- {
		saved := monk.networks[index]

		found := false
		for _, network := range networks {
			if same(saved, network) {
				found = true
				break
			}
		}

		if found {
			continue
		}

		monk.networks = append(
			monk.networks[:index],
			monk.networks[index+1:]...,
		)

		logger.Infof(
			"network removed: %s dev %s",
			saved.BroadcastAddress,
			saved.Interface,
		)
	}

	monk.networks = append(monk.networks, toAdd...)
}

func (monk *Monk) getPeer(id string) (*Peer, error) {
	var found *Peer
	peers := monk.peers.get()
	for i, _ := range peers {
		peer := peers[i]

		if peer.Machine == id {
			if found != nil {
				return nil, karma.
					Describe(found.IP, found.Fingerprint.String()).
					Describe(peer.IP, peer.Fingerprint.String()).
					Format(
						nil,
						"found two peers with the same machine id",
					)
			}

			found = &peer
		}
	}

	if found == nil {
		return nil, fmt.Errorf("no such peer: %s", id)
	}

	return found, nil
}
