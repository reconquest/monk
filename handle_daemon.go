package main

import (
	"os"
	"strconv"
	"sync"
	"syscall"

	"github.com/reconquest/sign-go"
)

func handleDaemon(args map[string]interface{}) {
	var (
		port, _           = strconv.Atoi(args["--port"].(string))
		allowedInterfaces = args["--interface"].([]string)
		socketPath        = args["--socket"].(string)
		dataDir           = args["--data-dir"].(string)
	)

	err := ensureDataDir(dataDir)
	if err != nil {
		fatalh(err, "unable to ensure data dir")
	}

	security, err := getSecureLayer(dataDir, true)
	if err != nil {
		fatalh(err, "unable to ensure TLS certificate exists")
	}

	machineID, err := getMachineID()
	if err != nil {
		fatalh(err, "unable to get machine id")
	}

	monk := NewMonk(
		machineID,
		security,
		dataDir,
		port,
		argInt(args, "--stream-buffer-size"),
	)

	err = monk.bind()
	if err != nil {
		fatalln(err)
	}

	socket, err := initSocket(socketPath, monk.HandleSock)
	if err != nil {
		fatalh(
			err,
			"unable to initialize unix sock: %s",
			socketPath,
		)
	}

	tcp, err := initTCP(
		port,
		security,
		monk.HandleTCPPacket,
		monk.HandleTCPConnection,
	)
	if err != nil {
		fatalh(err, "unable to initialize tcp at: %s", port)
	}

	monk.SetNetworks(
		filterNetworks(
			getNetworks(),
			allowedInterfaces,
		),
	)

	go watchNetworks(monk, allowedInterfaces)

	go monk.broadcastPresence()
	go monk.observe()

	go sign.Notify(func(os.Signal) bool {
		err := socket.Close()
		if err != nil {
			errorh(err, "unable to gracefully stop listening unix socket")
		}

		err = tcp.Close()
		if err != nil {
			errorh(err, "unable to gracefully stop listening tcp")
		}

		monk.Close()

		return false
	}, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	withWait(socket.Serve, tcp.Serve)
}

func withWait(fns ...func()) {
	workers := &sync.WaitGroup{}
	for _, fn := range fns {
		workers.Add(1)
		go func(fn func()) {
			defer workers.Done()
			fn()
		}(fn)

	}
	workers.Wait()
}
