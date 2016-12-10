package main

import (
	"net"
	"strconv"
	"time"

	"github.com/godbus/dbus"
	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
	"github.com/reconquest/colorgful"
)

const (
	dbusInterface = "com.github.reconquest.monk"
	dbusPath      = "/com/github/reconquest/monk"
)

var (
	version = "[manual build]"
	usage   = "monk " + version + `

blah

Usage:
  monk [options]
  monk -h | --help
  monk --version

Options:
  -p --port <port>           Specify port [default: 12345].
  -h --help                  Show this screen.
  --version                  Show version.
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

	logger.SetFormat(
		colorgful.MustApplyDefaultTheme(
			"${time} ${level:%s:left} ${prefix}%s",
			colorgful.Default,
		),
	)

	logger.SetLevel(lorg.LevelDebug)

	var (
		port, _ = strconv.Atoi(args["--port"].(string))
		//minConnections, _ = strconv.Atoi(args["--min-connections"].(string))
		//maxConnections, _ = strconv.Atoi(args["--max-connections"].(string))
	)

	monk := NewMonk(port)

	var err error
	monk.dbus, err = dbus.SessionBus()
	if err != nil {
		fatalh(err, "can't create dbus session")
	}

	err = monk.bind()
	if err != nil {
		fatalln(err)
	}

	for _, network := range getNetworks() {
		if network.IP.To4() == nil {
			continue
		}

		monk.addNetwork(Network{network})
	}

	time.Sleep(time.Second)

	go monk.observe()

	go func() {
		for range time.Tick(heartbeatInterval) {
			monk.heartbeat()
		}
	}()

	select {}
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
