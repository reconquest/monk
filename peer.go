package main

import (
	"net"
	"sync"
	"time"
)

type Peer struct {
	ip       net.IP
	networks []string
	last     time.Time
	peers    []*Peer
}

type Peers struct {
	peers []*Peer
	mutex *sync.Mutex
}

func (peers *Peers) add(peer *Peer) {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	peers.peers = append(peers.peers, peer)
}

func (peers *Peers) remove(peer *Peer) {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	for index, added := range peers.peers {
		if added == peer {
			peers.peers = append(
				peers.peers[:index],
				peers.peers[index+1:]...,
			)

			break
		}
	}
}

func (peers *Peers) cleanup(timeout time.Duration) {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	clean := []*Peer{}
	for _, peer := range peers.peers {
		if time.Now().Sub(peer.last) > timeout {
			continue
		}

		clean = append(clean, peer)
	}

	peers.peers = clean
}

func (peers *Peers) len() int {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	return len(peers.peers)
}

func (peers *Peers) getNetworks() []string {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	networks := []string{}

	for _, peer := range peers.peers {
		for _, network := range peer.networks {
			found := false
			for _, already := range networks {
				if already == network {
					found = true
					break
				}
			}

			if !found {
				networks = append(networks, network)
			}
		}
	}

	return networks
}
