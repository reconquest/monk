package main

import (
	"encoding/binary"
	"net"
	"strings"
	"time"
)

const (
	intervalWatchNetworks = time.Second * 5
)

type Network struct {
	Interface        string
	BroadcastAddress net.IP
	*net.IPNet
}

type Networks []Network

func (networks Networks) Strings() []string {
	strings := []string{}
	for _, network := range networks {
		strings = append(strings, network.String())
	}
	return strings
}

func (network Network) String() string {
	return network.IPNet.String()
}

func getBroadcastAddress(ipnet *net.IPNet) net.IP {
	ip4 := ipnet.IP.To4()
	mask4 := net.IP(ipnet.Mask).To4()

	ip := make(net.IP, len(ip4))
	binary.BigEndian.PutUint32(
		ip,
		binary.BigEndian.Uint32(ip4)|^binary.BigEndian.Uint32(mask4),
	)
	return ip
}

func watchNetworks(monk *Monk, allowedInterfaces []string) {
	for {
		time.Sleep(intervalWatchNetworks)

		networks := filterNetworks(getNetworks(), allowedInterfaces)
		monk.SetNetworks(networks)
	}
}

func filterNetworks(networks []Network, allowedInterfaces []string) []Network {
	filtered := []Network{}
	for _, network := range networks {
		if len(allowedInterfaces) == 0 {
			filtered = append(filtered, network)
			continue
		}

		for _, allowed := range allowedInterfaces {
			if strings.HasPrefix(network.Interface, allowed) {
				filtered = append(filtered, network)
				break
			}
		}
	}
	return filtered
}

func getNetworks() []Network {
	networks := []Network{}

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addresses, err := iface.Addrs()
		if err != nil {
			errorh(
				err,
				"unable to get addresses of interface: %s",
				iface.Name,
			)
			continue
		}

		for _, address := range addresses {
			ipnet := address.(*net.IPNet)
			if ipnet.IP.IsLoopback() {
				continue
			}

			if ipnet.IP.To4() == nil {
				continue
			}

			networks = append(
				networks,
				Network{
					Interface:        iface.Name,
					BroadcastAddress: getBroadcastAddress(ipnet),
					IPNet:            ipnet,
				},
			)
		}
	}

	return networks
}

func inSlice(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}

	return false
}
