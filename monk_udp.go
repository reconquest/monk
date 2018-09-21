package main

import (
	"bytes"
	"net"
	"time"
)

func (monk *Monk) bind() error {
	udp, err := net.ListenUDP("udp", &net.UDPAddr{Port: monk.port})
	if err != nil {
		return err
	}

	logger.Infof("listening at :%d", monk.port)

	monk.udp = udp

	return nil
}

func (monk *Monk) observe() {
	for {
		remote, packet, err := monk.readBroadcast()
		if err != nil {
			errorh(err, "unable to read packet")

			continue
		}

		go monk.handle(remote.(*net.UDPAddr), packet)
	}
}

func (monk *Monk) getNetworks() []Network {
	monk.mutex.Lock()
	defer monk.mutex.Unlock()

	networks := make([]Network, len(monk.networks))
	copy(networks, monk.networks)

	return networks
}

func (monk *Monk) broadcastPresence() {
	packet := PacketPresence{
		ID:          monk.machine,
		Fingerprint: monk.security.Fingerprint,
	}

	for {
		networks := monk.getNetworks()
		for _, network := range networks {
			monk.broadcast(network, packet)
		}

		time.Sleep(intervalPresence)
	}
}

func (monk *Monk) readBroadcast() (net.Addr, []byte, error) {
	packet := make([]byte, 1024*4)

	size, remote, err := monk.udp.ReadFrom(packet)
	if err != nil {
		return nil, nil, err
	}

	return remote, packet[:size], nil
}

func (monk *Monk) handle(remote *net.UDPAddr, data []byte) {
	myIP := false
	for _, network := range monk.networks {
		if bytes.Compare(remote.IP, network.IP) == 0 {
			myIP = true
			break
		}
	}

	packet, err := unpack(data)
	if err != nil {
		errorh(
			err,
			"unable to unpack packet: %s", data,
		)
		return
	}

	var presence PacketPresence
	err = packet.Bind(&presence)
	if err != nil {
		errorh(err, "unable to bind packet: %s", packet.Signature)
		return
	}

	if myIP && presence.ID == monk.machine {
		return
	}

	var latency time.Duration
	if presence.At.IsZero() {
		latency = time.Duration(0)
	} else {
		latency = time.Since(presence.At)
	}

	updated := monk.peers.updateLastSeen(
		remote.IP.String(),
		presence.Fingerprint,
		presence.ID,
		latency,
	)
	if updated {
		logger.Debugf(
			"presence: %s %s %s %v",
			presence.ID,
			remote.IP,
			presence.Fingerprint,
			latency,
		)
	} else {
		peer := Peer{
			IP:          remote.IP.String(),
			Machine:     presence.ID,
			Fingerprint: presence.Fingerprint,
			LastSeen:    time.Now(),
			Latency:     latency,
		}

		monk.peers.add(peer)

		logger.Infof(
			"new monk: %s %s %s %v",
			presence.ID,
			remote.IP.String(),
			presence.Fingerprint,
			latency,
		)
	}

	return
}

func (monk *Monk) broadcast(network Network, data Packetable) {
	debugf(
		"broadcasting to %s:%d dev %s",
		network.BroadcastAddress,
		monk.port,
		network.Interface,
	)

	_, err := monk.udp.WriteTo(
		pack(data),
		&net.UDPAddr{IP: network.BroadcastAddress, Port: monk.port},
	)
	if err != nil {
		errorh(
			err,
			"unable to send packet to %s:%d dev %s",
			network.BroadcastAddress, monk.port, network.Interface,
		)
		return
	}
}
