package main

import (
	"bytes"
	"sync"
	"time"
)

type Peer struct {
	IP          string      `json:"ip"`
	Trusted     bool        `json:"trusted"`
	Machine     string      `json:"machine"`
	Fingerprint Fingerprint `json:"fingerprint"`
	LastSeen    time.Time   `json:"last_seen"`
	data        map[string]interface{}
}

type Peers struct {
	peers []Peer
	mutex *sync.Mutex
}

func (peer *Peer) String() string {
	return peer.Machine + " " + peer.IP + " " + peer.Fingerprint.String()
}

func (peers *Peers) add(peer Peer) {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	peers.peers = append(peers.peers, peer)
}

func (peers *Peers) updateLastSeen(
	ip string,
	fingerprint Fingerprint,
	id string,
) bool {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	for i, peer := range peers.peers {
		if peer.IP == ip &&
			bytes.Compare(fingerprint, peer.Fingerprint) == 0 &&
			peer.Machine == id {
			peers.peers[i].LastSeen = time.Now()
			return true
		}
	}

	return false
}

func (peers *Peers) remove(peer Peer) {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	for index, known := range peers.peers {
		if peer.IP == known.IP &&
			bytes.Compare(peer.Fingerprint, known.Fingerprint) == 0 &&
			peer.Machine == known.Machine {
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

	clean := []Peer{}
	for _, peer := range peers.peers {
		if time.Now().Sub(peer.LastSeen) > timeout {
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

func (peers *Peers) get() []Peer {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	data := make([]Peer, len(peers.peers))
	copy(data, peers.peers)
	return data
}
