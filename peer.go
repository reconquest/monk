package main

import (
	"net"
	"sync"
	"time"
)

type Peer struct {
	ip      net.IP
	network string
	last    time.Time
	data    map[string]interface{}
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

func (peers *Peers) find(ip net.IP, network string) (*Peer, bool) {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	for _, peer := range peers.peers {
		if peer.ip.String() == ip.String() && peer.network == network {
			return peer, true
		}
	}

	return nil, false
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
