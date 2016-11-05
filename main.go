package main

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
	"github.com/reconquest/colorgful"
)

var (
	exit = os.Exit
)

var (
	version = "[manual build]"
	usage   = "cure " + version + `

blah

Usage:
  cure [options]
  cure -h | --help
  cure --version

Options:
  -p --port <port>  Specify port [default: 12345].
  -h --help         Show this screen.
  --version         Show version.
`
)

var (
	heartbeatInterval = time.Millisecond * 300
)

var (
	logger = lorg.NewLog()
)

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	var (
		port, _ = strconv.Atoi(args["--port"].(string))
	)

	logger.SetFormat(
		colorgful.MustApplyDefaultTheme(
			"${time} ${level:%s:left} ${prefix}%s",
			colorgful.Default,
		),
	)

	logger.SetLevel(lorg.LevelDebug)

	peers := []*Peer{}
	for _, network := range getNetworks() {
		if network.IP.To4() == nil {
			continue
		}

		peer := NewPeer(
			network,
			port,
			logger.NewChildWithPrefix(""+network.String()+":"),
		)

		err := peer.connect()
		if err != nil {
			fatalh(err, "unable to establish connection using %s", network)
		}

		peers = append(peers, peer)
	}

	for _, peer := range peers {
		go peer.observe()
	}

	for _, peer := range peers {
		go func(peer *Peer) {
			for range time.Tick(heartbeatInterval) {
				peer.heartbeat()
			}
		}(peer)
	}

	select {}
}

func getBroadcastIP(network *net.IPNet) net.IP {
	ip := make(net.IP, len(network.IP.To4()))
	binary.BigEndian.PutUint32(
		ip,
		binary.BigEndian.Uint32(network.IP.To4())|^binary.BigEndian.Uint32(
			net.IP(network.Mask).To4(),
		),
	)
	return ip
}

func encode(message interface{}) []byte {
	encoded, _ := json.Marshal(message)
	return encoded
}

func getNetworks() []*net.IPNet {
	networks := []*net.IPNet{}

	addresses, _ := net.InterfaceAddrs()
	for _, address := range addresses {
		if address.(*net.IPNet).IP.IsLoopback() {
			continue
		}

		networks = append(networks, address.(*net.IPNet))
	}

	return networks
}
