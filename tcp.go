package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/reconquest/karma-go"
)

type TCPPacketHandler func(*TCPConnection, Packet) (Packetable, error)
type TCPConnectionHandler func(*TCPConnection) error

type TCPConnection struct {
	conn net.Conn
	peer *Peer
}

type TCP struct {
	listener         net.Listener
	security         *SecureLayer
	handlePacket     TCPPacketHandler
	handleConnection TCPConnectionHandler
}

func (server *TCP) Close() error {
	return server.listener.Close()
}

func initTCP(
	port int,
	security *SecureLayer,
	handlePacket TCPPacketHandler,
	handleConnection TCPConnectionHandler,
) (*TCP, error) {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	return &TCP{
		listener:         sock,
		security:         security,
		handleConnection: handleConnection,
		handlePacket:     handlePacket,
	}, nil
}

func (server *TCP) Serve() {
	for {
		fd, err := server.listener.Accept()
		if err != nil {
			break
		}

		go server.serve(fd)
	}
}

func (server *TCP) serve(fd net.Conn) {
	defer fd.Close()

	decoder := json.NewDecoder(fd)
	encoder := json.NewEncoder(fd)

	writeError := func(err error) {
		logger.Error(err)

		packet := PacketError{}
		packet.Error = "internal server error"

		errEncode := encoder.Encode(makePacket(packet))
		if errEncode != nil {
			errorh(errEncode, "tcp: unable to write query")
		}
	}

	conn := &TCPConnection{
		conn: fd,
	}

	for {
		var query Packet
		err := decoder.Decode(&query)
		if err != nil {
			if err == io.EOF {
				return
			}

			writeError(karma.Format(err, "tcp: unable to read message"))
			break
		}

		reply, err := server.handlePacket(conn, query)
		if err != nil {
			writeError(karma.Format(err, "tcp: unable to serve packet"))
			break
		}

		if reply != nil {
			err = encoder.Encode(makePacket(reply))
			if err != nil {
				writeError(karma.Format(err, "tcp: unable to write reply packet"))
				continue
			}

			if reply.Signature() == SignatureEncryptConnection {
				err := server.handleConnection(conn)
				if err != nil {
					logger.Error(err)
				}
				break
			}

		}
	}
}
