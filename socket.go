package main

import (
	"encoding/json"
	"io"
	"net"
	"os"

	"github.com/reconquest/karma-go"
)

type Socket struct {
	path     string
	listener net.Listener
	handler  PacketHandler
}

func (socket *Socket) Close() error {
	return socket.listener.Close()
}

func initSocket(path string, handler PacketHandler) (*Socket, error) {
	sock, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	return &Socket{
		path:     path,
		listener: sock,
		handler:  handler,
	}, nil
}

func (socket *Socket) Serve() {
	defer os.Remove(socket.path)

	for {
		fd, err := socket.listener.Accept()
		if err != nil {
			break
		}

		go socket.serve(fd)
	}
}

func (socket *Socket) serve(fd net.Conn) {
	defer fd.Close()

	decoder := json.NewDecoder(fd)
	encoder := json.NewEncoder(fd)

	writeError := func(err error) {
		logger.Error(err)

		packet := PacketError{}
		packet.Error = err.Error()

		errEncode := encoder.Encode(makePacket(packet))
		if errEncode != nil {
			errorh(errEncode, "sock: unable to write query")
		}
	}

	for {
		var query Packet
		err := decoder.Decode(&query)
		if err != nil {
			if err == io.EOF {
				return
			}

			writeError(karma.Format(err, "sock: unable to read message"))
			break
		}

		reply, err := socket.handler(query)
		if err != nil {
			writeError(karma.Format(err, "sock: unable to serve packet"))
			break
		}

		if reply == nil {
			panic("handler returned packet with empty signature")
		}

		err = encoder.Encode(makePacket(reply))
		if err != nil {
			writeError(karma.Format(err, "sock: unable to write reply packet"))
		}
	}
}
