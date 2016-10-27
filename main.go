package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
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
  -p --port <port>  Specify port.
  -h --help        Show this screen.
  --version       Show version.
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
	Address string
	Clock   int64
}

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	fmt.Printf("XXXXXX main.go:47 args: %#v\n", args)

	listenAddresses := []*net.UDPAddr{}

	for _, ip := range getNonLocalIPs() {
		if ip.To4() == nil {
			continue
		}

		listenAddresses = append(listenAddresses, &net.UDPAddr{
			IP:   ip,
			Port: 12345,
		})
	}

	//dpMessages := make(chan []byte, 0)

	waitGroup := sync.WaitGroup{}
	for _, address := range listenAddresses {
		waitGroup.Add(1)

		go func(address *net.UDPAddr) {
			connection, err := net.ListenUDP("udp4", address)
			if err != nil {
				panic(err)
			}

			go func() {
				for {
					broadcast(connection)
					time.Sleep(time.Second)
				}

			}()

			for {
				fmt.Printf("XXXXXX main.go:87 11: %#v\n", 11)
				buffer := make([]byte, 1024)
				connection.ReadFromUDP(buffer)
				log.Printf("main.go:45 %#v", buffer)
			}
		}(address)
	}

	waitGroup.Wait()
}

func broadcast(connection *net.UDPConn) error {
	udpAddr := connection.LocalAddr().(*net.UDPAddr)

	addr := getBroadcastAddress(
		&net.IPNet{
			IP:   udpAddr.IP,
			Mask: udpAddr.IP.DefaultMask(),
		},
	)

	message := PresenceMessage{}
	message.Address = udpAddr.String()
	message.Clock = time.Now().UnixNano()

	fmt.Printf("XXXXXX main.go:111 addr: %s\n", addr)

	connection.WriteToUDP(encode(message), &net.UDPAddr{
		IP:   addr,
		Port: udpAddr.Port,
	})

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

func getNonLocalIPs() []net.IP {
	nonLocalIPs := []net.IP{}

	addresses, _ := net.InterfaceAddrs()
	for _, address := range addresses {
		if address.(*net.IPNet).IP.IsLoopback() {
			continue
		}

		nonLocalIPs = append(nonLocalIPs, address.(*net.IPNet).IP)
	}

	return nonLocalIPs
}
