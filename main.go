package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/kovetskiy/godocs"
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
	heartbeatInterval = 50
)

type Peer struct {
	Addr      net.IPAddr
	Port      int
	Source    net.IPAddr
	Timestamp time.Time
}

type PresenceMessage struct {
	Date int64 `json:"date"`
}

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	var (
		port, _ = strconv.Atoi(args["--port"].(string))
	)

	workers := &sync.WaitGroup{}
	for _, network := range getNetworks() {
		if network.IP.To4() == nil {
			continue
		}

		workers.Add(1)
		go serve(network, port)
	}

	workers.Wait()
}

func serve(network *net.IPNet, port int) {
	address := &net.UDPAddr{
		IP:   network.IP,
		Port: port,
	}

	connection, err := net.ListenPacket("udp", address.String())
	if err != nil {
		panic(err)
	}

	go func() {
		address := getBroadcastAddress(network)

		for {
			err = broadcast(connection, address)
			if err != nil {
				panic(err)
			}

			time.Sleep(time.Second)
		}
	}()

	for {
		buffer := make([]byte, 1024)
		length, remote, err := connection.ReadFrom(buffer[:])
		if err != nil {
			panic(err)
		}

		fmt.Printf("XXXXXX %s: %s\n", remote, buffer[:length])
	}
}

func broadcast(connection net.PacketConn, address net.IP) error {
	message := PresenceMessage{}
	message.Date = time.Now().UnixNano()

	fmt.Printf("XXXXXX brcast: %v\n", message)

	_, err := connection.WriteTo(encode(message), &net.UDPAddr{
		IP:   address,
		Port: connection.LocalAddr().(*net.UDPAddr).Port,
	})
	if err != nil {
		return err
	}

	return nil
}

func getBroadcastAddress(network *net.IPNet) net.IP {
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
